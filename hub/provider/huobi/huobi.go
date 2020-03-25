package huobi

import (
	"bitcoin-kline/config"
	"bitcoin-kline/constant"
	"bitcoin-kline/logger"
	"bitcoin-kline/model"
	"encoding/json"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/parnurzeal/gorequest"
	"github.com/pkg/errors"
)

// 官网：https://www.hbg.com/zh-cn/
// 官方文档：https://huobiapi.github.io/docs/spot/v1/cn/#2d07d50c97
// 注：api-aws.huobi.pro域名对使用aws云服务的用户做了一定的链路延迟优化。
// 限频：10秒100次
// api接口：https://huobiapi.github.io/docs/spot/v1/cn/#ticker

type Provider struct {
	readChan map[string]chan *model.Kline

	breakMainLogic chan bool // 结束命令管道
	sync.WaitGroup
}

type Ticker struct {
	Id    int64   `json:"id"`
	Close float64 `json:"close"` // 本阶段最新价
	Open  float64 `json:"open"`  // 本阶段开盘价
	High  float64 `json:"high"`  // 本阶段最高价
	Low   float64 `json:"low"`   // 本阶段最低价
	//Amount float64    `json:"amount"` // 以基础币种计量的交易量
	Count int     `json:"count"` // 交易次数
	Vol   float64 `json:"vol"`   // 以报价币种计量的交易量
	//Ask    []string `json:"ask"`    // 当前的最低卖价 [price, quote volume]
	//Bid    []string `json:"bid"`    // 当前的最高买价 [price, quote volume]
}

type ApiResponse struct {
	Status   string  `json:"status"`
	ErrorMsg string  `json:"err-msg"`
	Ticker   *Ticker `json:"tick"`
}

var (
	huobiCoinMap = map[string]string{
		constant.CoinTypeETHUSDT: "ethusdt",
		constant.CoinTypeBTCUSDT: "btcusdt",
	}
)

func NewProvider() *Provider {
	p := &Provider{
		readChan:       make(map[string]chan *model.Kline),
		breakMainLogic: make(chan bool),
	}

	for _, coinTyp := range config.SupportCoinTypes {
		if _, ok := huobiCoinMap[coinTyp]; ok {
			p.readChan[coinTyp] = make(chan *model.Kline)
		}
	}

	return p
}

func (p *Provider) ReadChan(coinType string) <-chan *model.Kline {
	return p.readChan[coinType]
}

func (p *Provider) StartCollect() {
	for _, coinType := range config.SupportCoinTypes {
		if _, ok := huobiCoinMap[coinType]; ok {
			p.Add(1)
			go func(c string) {
				defer p.Done()
				p.loop(c)
			}(coinType)
		}
	}
}

func (p *Provider) loop(params ...interface{}) {
	if len(params) <= 0 {
		return
	}
	coinType := params[0].(string)
	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		select {
		case <-t.C:

			ticker, err := p.getTicker(coinType)
			if err != nil {
				logger.Error("HuobiProvider_getTicker", coinType, err.Error())
				break
			}
			now := time.Now().Unix()
			price := strconv.FormatFloat(ticker.Close, 'f', 4, 64)
			vol := strconv.FormatFloat(ticker.Vol, 'f', 4, 64)

			item := &model.Kline{
				CoinType:    coinType,
				High:        price,
				Low:         price,
				Open:        price,
				Close:       price,
				CreateTime:  now,
				UpdateTime:  now,
				TimeScale:   "1s",
				Origin:      constant.ProviderHuoBiOriginType,
				OriginPrice: "",
				Volume:      vol,
			}

			p.handleKline(coinType, item)
			break
		case <-p.breakMainLogic:
			return
		}
	}
}

func (p *Provider) handleKline(coinType string, kline *model.Kline) {
	select {
	case p.readChan[coinType] <- kline:

	case <-time.After(time.Second * constant.ProviderDataExpireTime):
	case <-p.breakMainLogic:
	}
}

func (p *Provider) Stop() {
	close(p.breakMainLogic)
	p.Wait()
}

// request data here
// 此接口获取ticker信息同时提供最近24小时的交易聚合信息。
func (p *Provider) getTicker(coinType string) (*Ticker, error) {
	symbol := huobiCoinMap[coinType]
	request := gorequest.New()
	if os.Getenv("RUNMODE") == "dev" || config.CURMODE == "dev" {
		request = request.Proxy("socks5://127.0.0.1:1088")
	}
	_, body, errs := request.Get("https://api-aws.huobi.pro/market/detail/merged?symbol=" + symbol).Timeout(3 * time.Second).End()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	resp := &ApiResponse{}
	if err := json.Unmarshal([]byte(body), resp); err != nil {
		return nil, err
	}
	if resp.Status != "ok" {
		return nil, errors.New(resp.ErrorMsg)
	}

	tick := resp.Ticker
	if tick == nil {
		return nil, errors.New("data invalid")
	}

	return tick, nil
}

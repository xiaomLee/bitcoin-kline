package zb

import (
	"bitcoin-kline/config"
	"bitcoin-kline/constant"
	"bitcoin-kline/logger"
	"bitcoin-kline/model"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/parnurzeal/gorequest"
	"github.com/pkg/errors"
)

// 官网：https://www.zbex3.com/
// Api文档：https://www.zb.com/api
// 接口限流：单个API KEY限制每秒钟60次访问，一秒钟内60次以上的请求，将会视作无效。
// https://www.zb.com/api#API%E4%BB%8B%E7%BB%8DAPI%E6%8E%A5%E5%8F%A3%E8%AF%B4%E6%98%8E
// ticker接口：https://www.zb.com/api#pmtjkvevzyqinkb

type Provider struct {
	readChan map[string]chan *model.Kline

	breakMainLogic chan bool // 结束命令管道
	sync.WaitGroup
}

type Ticker struct {
	High string `json:"high"` // 最高价
	Low  string `json:"low"`  // 最低价
	Last string `json:"last"` // 最新成交价
	Sell string `json:"sell"` // 卖一价
	Buy  string `json:"buy"`  // 买一价
	Vol  string `json:"vol"`  // 成交量(最近的24小时)
}

type ApiResponse struct {
	Date   string  `json:"date"`
	Ticker *Ticker `json:"ticker"`
	Error  string  `json:"error"`
}

var (
	zbCoinMap = map[string]string{
		constant.CoinTypeETHUSDT: "eth_usdt",
		constant.CoinTypeBTCUSDT: "btc_usdt",
	}
)

func NewZbProvider() *Provider {
	p := &Provider{
		readChan:       make(map[string]chan *model.Kline),
		breakMainLogic: make(chan bool),
	}

	for _, coinTyp := range config.SupportCoinTypes {
		p.readChan[coinTyp] = make(chan *model.Kline)
	}

	return p
}

func (p *Provider) ReadChan(coinType string) <-chan *model.Kline {
	return p.readChan[coinType]
}

func (p *Provider) StartCollect() {
	for _, coinType := range config.SupportCoinTypes {
		// 不支持的交易对跳过
		if _, ok := zbCoinMap[coinType]; ok {
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
			kline, err := p.getTicker(coinType)
			if err != nil || kline == nil {
				logger.Error("ZbProvider_getTicker", coinType, err.Error())
				break
			}
			now := time.Now().Unix()

			item := &model.Kline{
				CoinType:    coinType,
				High:        kline.Last,
				Low:         kline.Last,
				Open:        kline.Last,
				Close:       kline.Last,
				CreateTime:  now,
				UpdateTime:  now,
				TimeScale:   "1s",
				Origin:      constant.ProviderZBOriginType,
				OriginPrice: "",
				Volume:      kline.Vol,
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

// request data here 获取最新价
func (p *Provider) getTicker(coinType string) (*Ticker, error) {
	zbCoinType := zbCoinMap[coinType]
	request := gorequest.New()
	request = request.AppendHeader("User-Agent", "Chrome/39.0.2171.71")
	if os.Getenv("RUNMODE") == "dev" || config.CURMODE == "dev" {
		request = request.Proxy("socks5://127.0.0.1:1088")
	}

	_, body, errs := request.Get("http://api.zb.cn/data/v1/ticker?market=" + zbCoinType).Timeout(3 * time.Second).End()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	resp := &ApiResponse{}
	if err := json.Unmarshal([]byte(body), resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}

	tick := resp.Ticker
	if tick == nil {
		return nil, errors.New("data invalid")
	}

	return tick, nil
}

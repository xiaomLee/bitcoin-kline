package bitz

import (
	"bitcoin-kline/config"
	"bitcoin-kline/constant"
	"bitcoin-kline/logger"
	"bitcoin-kline/model"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/parnurzeal/gorequest"
)

// 官网：https://www.bitz.top/ www.bit-z.pro www.bit-z.com www.bitz.com
// API文档：https://apidoc.bitz.top/cn/
// 限频：
// 行情接口：https://apidoc.bitz.top/cn/market-quotation-data/Get-ticker-data.html

type Provider struct {
	readChan map[string]chan *model.Kline

	//currentKline map[string]chan *model.Kline

	breakMainLogic chan bool // 结束命令管道
	sync.WaitGroup
}

type Ticker struct {
	Last string `json:"now"`    // 本阶段最新价
	High string `json:"high"`   // 本阶段最高价
	Low  string `json:"low"`    // 本阶段最低价
	Open string `json:"open"`   // 本阶段开盘价
	Vol  string `json:"volume"` // 以报价币种计量的交易量
}

type ApiResponse struct {
	Status  int     `json:"status"`
	Message string  `json:"msg"`
	Data    *Ticker `json:"data"`
}

var (
	bitzCoinMap = map[string]string{
		constant.CoinTypeETHUSDT: "eth_usdt",
		constant.CoinTypeBTCUSDT: "btc_usdt",
	}
)

func NewProvider() *Provider {
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
		if _, ok := bitzCoinMap[coinType]; ok {
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
				logger.Error("BitzProvider_getTicker", nil, err.Error())
				break
			}
			now := time.Now().Unix()
			price := ticker.Last
			vol := ticker.Vol

			item := &model.Kline{
				CoinType:    coinType,
				High:        price,
				Low:         price,
				Open:        price,
				Close:       price,
				CreateTime:  now,
				UpdateTime:  now,
				TimeScale:   "1s",
				Origin:      constant.ProviderBitzOriginType,
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
	symbol := bitzCoinMap[coinType]
	request := gorequest.New()
	request = request.AppendHeader("User-Agent", "Chrome/39.0.2171.71")
	url := fmt.Sprintf("https://apiv2.bitz.com/Market/ticker?symbol=%s", symbol)
	_, body, errs := request.Get(url).Timeout(5 * time.Second).End()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	resp := &ApiResponse{}
	if err := json.Unmarshal([]byte(body), resp); err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, errors.New(resp.Message)
	}

	tick := resp.Data
	if tick == nil {
		return nil, errors.New("data invalid")
	}

	return tick, nil
}

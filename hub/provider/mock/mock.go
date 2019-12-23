package mock

import (
	"bitcoin-kline/config"
	"bitcoin-kline/constant"
	"bitcoin-kline/model"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

// 本地mock 开发测试使用
type Provider struct {
	readChan       map[string]chan *model.Kline
	breakMainLogic chan bool // 结束命令管道
	sync.WaitGroup
}

type Ticker struct {
	High   string `json:"high"`
	Open   string `json:"open"`
	Low    string `json:"low"`
	Close  string `json:"close"`
	Sell   string `json:"sell"`
	Buy    string `json:"buy"`
	Volume string `json:"volume"`
}

var current = 100.00

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
		p.Add(1)
		go func() {
			defer p.Done()
			p.loop(coinType)
		}()
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
			ticker, _ := p.getTicker(coinType)
			now := time.Now().Unix()
			item := &model.Kline{
				CoinType:    coinType,
				High:        ticker.Close,
				Low:         ticker.Close,
				Open:        ticker.Close,
				Close:       ticker.Close,
				CreateTime:  now,
				UpdateTime:  now,
				TimeScale:   "1s",
				Origin:      constant.ProviderMockOriginType,
				OriginPrice: ticker.Close,
				Volume:      ticker.Volume,
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

// mock data here
func (p *Provider) getTicker(coinType string) (*Ticker, error) {
	op := rand.Intn(21) - 10
	current = current + rand.Float64()*float64(op)/10
	price := strconv.FormatFloat(current, 'f', 4, 64)
	vol := strconv.Itoa(rand.Intn(2000) + 1000)
	ticker := Ticker{
		High:   price,
		Open:   price,
		Low:    price,
		Close:  price,
		Volume: vol,
	}

	return &ticker, nil
}

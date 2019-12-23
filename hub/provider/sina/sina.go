package sina

import (
	"bitcoin-kline/config"
	"bitcoin-kline/constant"
	"bitcoin-kline/model"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/parnurzeal/gorequest"
)

// 橡胶期货
// 官网：https://finance.sina.com.cn/futures/quotes/RU0.shtml
// Api文档：http://joeychou.me/blog/53.html
// ticker接口：https://www.zb.com/api#pmtjkvevzyqinkb

type Provider struct {
	readChan map[string]chan *model.Kline

	breakMainLogic chan bool // 结束命令管道
	sync.WaitGroup
}

type Ticker struct {
	High string `json:"high"` // 最高价
	Open string `json:"high"` // 开盘价
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
	coinMap = map[string]string{
		constant.CoinTypeRUCNY: "nf_RU0",
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
		// 不支持的交易对跳过
		if _, ok := coinMap[coinType]; ok {
			p.Add(1)
			go func() {
				defer p.Done()
				p.loop(coinType)
			}()
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
				//				logger.Error("SinaProvider_getTicker", nil, err.Error())
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
				Origin:      constant.ProviderSinaOriginType,
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

// 数据返回为字符串：var hq_str_RU0="橡胶连续,225956,13145.00,13185.00,13005.00,13120.00,13115.00,13120.00,13120.00,13107.00,13265.00,1,6,434190,272890,沪,橡胶,2019-12-13,0,13455.000,13055.000,13455.000,12535.000,13455.000,11870.000,13455.000,11330.000,250.266";
// 0：豆粕连续，名字
// 1：145958，不明数字（难道是数据提供商代码？）
// 2：3170，开盘价
// 3：3190，最高价
// 4：3145，最低价
// 5：3178，昨日收盘价 （2013年6月27日）
// 6：3153，买价，即“买一”报价
// 7：3154，卖价，即“卖一”报价
// 8：3154，最新价，即收盘价
// 9：3162，结算价
// 10：3169，昨结算
// 11：1325，买 量
// 12：223，卖 量
// 13：1371608，持仓量
// 14：1611074，成交量
// 15：连，大连商品交易所简称
// 16：豆粕，品种名简称
// 17：2013-06-28，日期

// request data here 获取最新价
func (p *Provider) getTicker(coinType string) (*Ticker, error) {
	_, ok := coinMap[coinType]
	// 不支持的交易对直接返回nil
	if !ok {
		println("coinType not sup", coinType)
		return nil, nil
	}

	request := gorequest.New()
	request = request.AppendHeader("User-Agent", "Chrome/39.0.2171.71")
	request = request.AppendHeader("Accept-Encoding", "gzip, deflate")

	_, body, errs := request.Get("http://hq.sinajs.cn/list=nf_RU0").Timeout(3 * time.Second).End()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	data := strings.Split(body, ",")
	if len(data) < 18 {
		return nil, errors.New("body data err")
	}

	tick := &Ticker{
		Sell: data[7],
		Buy:  data[6],
		Vol:  data[14],
	}

	// 随机波动
	//current, _ := strconv.ParseFloat(data[8], 64)
	//op := rand.Intn(21) - 10
	//current = current + current*rand.Float64()*float64(op)/10000
	//last := strconv.FormatFloat(current, 'f', 4, 64)
	//
	//if ret, _ := common.BcCmp(last, "12000.0000"); ret <= 0 {
	//	fmt.Printf("now:%v currrent:%f, op:%d, last:%s \n", time.Now(), current, op, last)
	//	for i, val := range data {
	//		if i == 0 {
	//			continue
	//		}
	//		fmt.Printf("%d:%+v  ", i, val)
	//	}
	//	print("\n")
	//}

	last := data[8]

	if coinType == constant.CoinTypeRUCNY {
		tick.High = last
		tick.Open = last
		tick.Last = last
		tick.Low = last

		return tick, nil
	}

	tick.High = last
	tick.Open = last
	tick.Last = last
	tick.Low = last

	return tick, nil
}

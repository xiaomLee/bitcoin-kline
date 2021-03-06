package worker

import (
	"bitcoin-kline/common"
	"bitcoin-kline/config"
	"bitcoin-kline/constant"
	"bitcoin-kline/hub/provider"
	"bitcoin-kline/hub/provider/binance"
	"bitcoin-kline/hub/provider/bitmax"
	"bitcoin-kline/hub/provider/bitz"
	"bitcoin-kline/hub/provider/gateio"
	"bitcoin-kline/hub/provider/huobi"
	"bitcoin-kline/hub/provider/mock"
	"bitcoin-kline/hub/provider/okex"
	"bitcoin-kline/hub/provider/sina"
	"bitcoin-kline/hub/provider/zb"
	"bitcoin-kline/logger"
	"bitcoin-kline/model"
	"fmt"
	"math"
	"sort"
	"strconv"
	"sync"
	"time"
)

type ProviderWorker struct {
	providers    map[string]provider.Provider
	currentKline map[string]*model.Kline

	breakMainLogic chan bool // 结束命令管道

	sync.RWMutex
	sync.WaitGroup
}

var (
	providers     []string
	fixedDataChan map[string]chan *model.Kline
)

func InitProviderWorker() {
	fixedDataChan = make(map[string]chan *model.Kline)
	for _, coinType := range config.SupportCoinTypes {
		fixedDataChan[coinType] = make(chan *model.Kline)
	}

	providers = make([]string, 0)
	if config.CURMODE == config.ENV_DEV {
		providers = append(providers, constant.ProviderMock)
		//providers = append(providers, constant.ProviderSina)
	} else {
		//providers = append(providers, constant.ProviderMock)
		providers = append(providers, constant.ProviderZB)
		providers = append(providers, constant.ProviderHuoBi)
		providers = append(providers, constant.ProviderOkex)
		providers = append(providers, constant.ProviderBitz)
		providers = append(providers, constant.ProviderGateio)
		providers = append(providers, constant.ProviderBinance)
		providers = append(providers, constant.ProviderBitmax)
		providers = append(providers, constant.ProviderBitmax)
		//providers = append(providers, constant.ProviderSina) // 橡胶期货数据
	}
}

func NewProviderWorker() *ProviderWorker {
	p := &ProviderWorker{
		providers:      make(map[string]provider.Provider),
		currentKline:   make(map[string]*model.Kline),
		breakMainLogic: make(chan bool),
	}
	return p
}

func (w *ProviderWorker) Start() error {
	for _, val := range providers {
		switch val {
		case constant.ProviderMock:
			w.providers[val] = mock.NewProvider()
		case constant.ProviderZB:
			w.providers[val] = zb.NewZbProvider()
		case constant.ProviderHuoBi:
			w.providers[val] = huobi.NewProvider()
		case constant.ProviderOkex:
			w.providers[val] = okex.NewProvider()
		case constant.ProviderBitz:
			w.providers[val] = bitz.NewProvider()
		case constant.ProviderGateio:
			w.providers[val] = gateio.NewProvider()
		case constant.ProviderBinance:
			w.providers[val] = binance.NewProvider()
		case constant.ProviderBitmax:
			w.providers[val] = bitmax.NewProvider()
		case constant.ProviderSina:
			w.providers[val] = sina.NewProvider()

		}
	}

	// start collect data
	for _, p := range w.providers {
		p.StartCollect()
	}

	// 聚合修正多家供应商的数据
	for _, coinType := range config.SupportCoinTypes {
		w.Add(1)
		go func(c string) {
			defer w.Done()
			w.fixDataLoop(c)
		}(coinType)
	}

	return nil
}

// 结束主逻辑
func (w *ProviderWorker) Stop() {
	for _, p := range w.providers {
		p.Stop()
	}

	close(w.breakMainLogic)
	w.Wait()
}

func (w *ProviderWorker) fixDataLoop(params ...interface{}) {
	if len(params) == 0 {
		return
	}

	coinType := params[0].(string)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			items := w.readData(coinType)
			item := w.fixData(coinType, items)

			if item == nil {
				break
			}

			select {
			case fixedDataChan[coinType] <- item:
			case <-w.breakMainLogic:
			}

		case <-w.breakMainLogic:
			return
		}
	}
}

// 从数据商管道读取数据
func (w *ProviderWorker) readData(coinType string) []*model.Kline {
	items := make([]*model.Kline, 0)
	for _, p := range w.providers {
		select {
		case item := <-p.ReadChan(coinType):
			items = append(items, item)
			break
		default:
		}
	}

	return items
}

func (w *ProviderWorker) setCurrentKline(kline *model.Kline) {
	w.Lock()
	defer w.Unlock()
	w.currentKline[kline.CoinType] = kline
}

func (w *ProviderWorker) getCurrentKline(coinType string) *model.Kline {
	w.RLock()
	defer w.RUnlock()
	return w.currentKline[coinType]
}

func (w *ProviderWorker) fixData(coinType string, items []*model.Kline) *model.Kline {
	if len(items) == 0 {
		return nil
	}

	// 过滤异常值
	afterFilter := filterOutliers(items)
	prices := make([]map[string]string, 0)
	priceSum := "0"
	priceVol := "0"
	for _, item := range afterFilter {
		prices = append(prices, map[string]string{constant.ProviderOriginMap[item.Origin]: item.Close})
		priceSum, _ = common.BcAdd(priceSum, item.Close, 18)
		priceVol, _ = common.BcAdd(priceVol, item.Volume, 18)
	}

	// 计算市场平均值
	marketPrice, _ := common.BcDiv(priceSum, strconv.Itoa(len(afterFilter)), 4)
	vol, _ := common.BcDiv(priceVol, strconv.Itoa(len(afterFilter)), 4)

	// 构造kline
	now := time.Now().Unix()
	kline := &model.Kline{
		CoinType:    coinType,
		High:        marketPrice,
		Low:         marketPrice,
		Open:        marketPrice,
		Close:       marketPrice,
		CreateTime:  now,
		UpdateTime:  now,
		TimeScale:   "1s",
		Origin:      1,
		OriginPrice: marketPrice,
		Volume:      vol,
	}

	w.setCurrentKline(kline)
	return kline
}

// 虚盒法过滤异常值, 先从小到大排序
// https://baike.baidu.com/item/%E7%AE%B1%E5%BC%8F%E5%9B%BE
func filterOutliers(items []*model.Kline) []*model.Kline {
	length := len(items)
	if length < 4 {
		return items
	}
	sort.Slice(items, func(i, j int) bool {
		ret, _ := common.BcCmp(items[i].Close, items[j].Close)
		return ret < 0
	})

	val1, _ := common.BcMul("0.25", items[int(math.Floor(float64((length)/4)))].Close, 18)
	val2, _ := common.BcMul("0.75", items[int(math.Ceil(float64((length)/4)))].Close, 18)
	q1, _ := common.BcAdd(val1, val2, 18)

	val1, _ = common.BcMul("0.25", items[int(math.Floor(float64((length)*3/4)))].Close, 18)
	val2, _ = common.BcMul("0.75", items[int(math.Ceil(float64((length)*3/4)))].Close, 18)
	q3, _ := common.BcAdd(val1, val2, 18)

	iqr, _ := common.BcSub(q3, q1, 18)

	iqr15, _ := common.BcMul("1.5", iqr, 18)
	max, _ := common.BcAdd(q3, iqr15, 18)
	min, _ := common.BcSub(q1, iqr15, 18)

	result := make([]*model.Kline, 0)
	for i, item := range items {
		ret1, _ := common.BcCmp(item.Close, min)
		ret2, _ := common.BcCmp(item.Close, max)
		if ret1 < 0 || ret2 > 0 {
			logger.Error("providerworker_filterOutliers", items,
				fmt.Sprintf("provider:%s, close:%s, Q1:%s, Q3:%s, min:%s, max:%s index:%d", constant.ProviderOriginMap[item.Origin], item.Close, q1, q3, min, max, i))
			continue
		}
		result = append(result, item)
	}

	return result
}

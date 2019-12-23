package worker

import (
	"bitcoin-kline/config"
	"sync"
)

// 程序主逻辑
// read data from provider
// push data to dbworker		保存数据库
// push data to mqworker		推送k线
// push data to tickerworker	计算成交量

type KlineWorker struct {
	breakMainLogic chan bool // 结束命令管道
	sync.WaitGroup
}

func NewKlineWorker() *KlineWorker {
	kline := &KlineWorker{
		breakMainLogic: make(chan bool),
	}

	return kline
}

func (w *KlineWorker) Start() error {
	for _, coinType := range config.SupportCoinTypes {
		w.Add(1)
		go func() {
			defer w.Done()
			w.workLoop(coinType)
		}()
	}

	return nil
}

// 结束主逻辑
func (w *KlineWorker) Stop() {
	close(w.breakMainLogic)
	w.Wait()
}

// read data from provider into dataChan
func (w *KlineWorker) workLoop(params ...interface{}) {
	if len(params) == 0 {
		return
	}
	coinType := params[0].(string)

	for {
		select {
		case kline := <-fixedDataChan[coinType]:
			if kline == nil {
				break
			}

			// 保存行情数据
			select {
			case klineDbChan <- kline.Copy():
			case <-w.breakMainLogic:
			}

			// 推送mq
			select {
			case klineMqChan <- kline.Copy():
			case <-w.breakMainLogic:
			}

		case <-w.breakMainLogic:
			return
		}
	}
}

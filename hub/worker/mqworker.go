package worker

import (
	"bitcoin-kline/common"
	"bitcoin-kline/config"
	"bitcoin-kline/constant"
	"bitcoin-kline/model"
	"fmt"
	"sync"

	"github.com/streadway/amqp"

	json "github.com/json-iterator/go"
)

type MqWorker struct {
	currentKline   map[string]*model.Kline
	breakMainLogic chan bool // 结束命令管道
	sync.WaitGroup
}

type Message struct {
	AppId     string `json:"appId"`
	EventType string `json:"eventType"` // 订阅的事件
	Body      string `json:"body"`
}

const DefaultMqChanSize = 1024

var (
	klineMqChan chan model.Kline
	appId       string
	exchange    string
	routerKey   = "kline"
)

func InitMqWorker() {
	exchange = config.GetConfig("rabbit", "exchange")
	appId = config.GetConfig("rabbit", "appId")
	klineMqChan = make(chan model.Kline, DefaultMqChanSize)
	common.GetRabbitInstance().ExchangeDeclare(exchange)

	//test consumer
	common.GetRabbitInstance().QueueDeclare("kline", exchange, routerKey)
	common.GetRabbitInstance().Consume("test_callback", "kline", klineCallBack)
}

func klineCallBack(delivery amqp.Delivery) {
	fmt.Printf("testMqSubcriptEvent:%+v \n", string(delivery.Body))
}

func NewMqWorker() *MqWorker {
	return &MqWorker{
		currentKline:   make(map[string]*model.Kline),
		breakMainLogic: make(chan bool),
	}
}

func (w *MqWorker) Start() error {
	w.Add(1)
	go func() {
		defer w.Done()
		w.pushLoop()
	}()

	return nil
}

// 结束主逻辑
func (w *MqWorker) Stop() {
	close(w.breakMainLogic)
	w.Wait()
}

func (w *MqWorker) pushLoop() {
	for {
		select {
		case kline := <-klineMqChan:
			// 此处设置 origin=1 标记k线数据来自市场
			kline.Origin = 1
			kline.Volume = "0"

			// 填充24小时 open high low vol
			existKline := model.GetDbKline(kline.CoinType, kline.CreateTime, "1D")
			if existKline != nil {
				kline.Volume = existKline.Volume
				kline.Open = existKline.Open
				kline.High = existKline.High
				kline.Low = existKline.Low
			}

			// 更新currentKline
			w.currentKline[kline.CoinType] = &kline

			// 发送msg
			event := constant.MqEventTypeTick + kline.CoinType
			msgBody := struct {
				EventType string       `json:"eventType"`
				Data      *model.Kline `json:"data"`
			}{
				EventType: event,
				Data:      &kline,
			}
			bytes, _ := json.Marshal(msgBody)
			pushMq(event, string(bytes))

		case <-w.breakMainLogic:
			return

		}
	}
}

func pushMq(eventType string, msgBody string) {
	message := Message{
		AppId:     appId,
		EventType: eventType,
		Body:      msgBody,
	}
	body, _ := json.Marshal(message)
	common.GetRabbitInstance().PushTransientMessage(exchange, routerKey, body)
	//fmt.Printf("msgBody:%s \n", body)
}

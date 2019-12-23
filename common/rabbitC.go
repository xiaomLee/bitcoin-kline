package common

import (
	"bitcoin-kline/logger"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type RabbitC struct {
	conn *amqp.Connection
	ch   *amqp.Channel

	asyncPushChan chan Message

	exit   chan bool
	exited chan bool

	url           string
	consumeQueues []string

	afterConnected     func(r *RabbitC, ch *amqp.Channel) error
	newMessageCallBack []func(amqp.Delivery)

	sync.WaitGroup
}

// NewRabbitC returns a RabbitC instance
// url is format like 'amqp://account:password@ip:port/vHost'
//
// for example: amqp://gateway-ws:gateway-ws123@127.0.0.1:5672/gateway-ws
// account is 'gateway-ws', password is 'gateway-ws123', ip is '127.0.0.1', port is '5672', vHost is 'gateway-ws'
func NewRabbitC(url string, afterConnected func(r *RabbitC, ch *amqp.Channel) error, consumeQueues []string) *RabbitC {
	r := &RabbitC{
		exit:               make(chan bool),
		exited:             make(chan bool),
		asyncPushChan:      make(chan Message, 1024),
		url:                url,
		consumeQueues:      consumeQueues,
		afterConnected:     afterConnected,
		newMessageCallBack: make([]func(amqp.Delivery), 0),
	}

	return r
}

// SetAfterConnectedFunc is a callback func, will be called after connect rabbit success.
// This is a great time to create exchange and queue
func (r *RabbitC) SetAfterConnectedFunc(afterConnected func(r *RabbitC, ch *amqp.Channel) error) {
	r.afterConnected = afterConnected
}

// SetConsumeQueues specify queues for consumption
func (r *RabbitC) SetConsumeQueues(queues []string) {
	r.consumeQueues = queues
}

// SetNewMessageCallBack specify callback func to handle new consume message from rabbit
func (r *RabbitC) SetNewMessageCallBack(c func(amqp.Delivery)) {
	r.newMessageCallBack = append(r.newMessageCallBack, c)
}

// Start start RabbitC
func (r *RabbitC) Start() (err error) {
	if err = r.initRabbit(); err != nil {
		return
	}

	if r.afterConnected != nil {
		if err = r.afterConnected(r, r.ch); err != nil {
			return
		}
	}

	if err = r.tryConsume(); err != nil {
		return
	}

	go r.watchRabbit()

	return
}

// Close close RabbitC
func (r *RabbitC) Close() {
	close(r.exit)
	r.Wait()
}

type Message struct {
	Exchange     string
	Router       string
	DeliveryMode uint8
	Body         []byte
}

// PushTransientMessage push a transient message to rabbit
func (r *RabbitC) PushTransientMessage(exchange, router string, body []byte) {
	r.asyncPushChan <- Message{
		Exchange:     exchange,
		Router:       router,
		DeliveryMode: amqp.Transient,
		Body:         body,
	}
}

// PushTransientMessage push a persistent message to rabbit
func (r *RabbitC) PushPersistentMessage(exchange, router string, body []byte) {
	r.asyncPushChan <- Message{
		Exchange:     exchange,
		Router:       router,
		DeliveryMode: amqp.Persistent,
		Body:         body,
	}
}

func (r *RabbitC) initRabbit() (err error) {
	if r.conn, err = amqp.Dial(r.url); err != nil {
		return
	}

	if r.ch, err = r.conn.Channel(); err != nil {
		_ = r.conn.Close()
		return
	}

	return
}

func (r *RabbitC) tryConsume() error {
	for _, queue := range r.consumeQueues {
		dch, err := r.ch.Consume(queue, "", true, false, false, false, nil)
		if err != nil {
			return err
		}

		go r.consumePump(dch)
	}

	return nil
}

func (r *RabbitC) consumePump(dch <-chan amqp.Delivery) {
	r.Add(1)

LOOP:
	for {
		select {
		case delivery, ok := <-dch:

			if !ok {
				break LOOP
			}

			for _, f := range r.newMessageCallBack {
				f(delivery)
			}
		}
	}

	r.Done()
}

func (r *RabbitC) watchRabbit() {
	notifyCloseChan := make(chan *amqp.Error)
	notifyCloseChan = r.conn.NotifyClose(notifyCloseChan)

	notifyReturnChan := make(chan amqp.Return)
	notifyReturnChan = r.ch.NotifyReturn(notifyReturnChan)

	reconnect := time.NewTicker(3 * time.Second)
	brokePipLine := false

	for {
		select {
		case <-r.exit:
			goto FINISH

		case m := <-r.asyncPushChan:
			_ = r.ch.Publish(m.Exchange, m.Router, false, false, amqp.Publishing{
				DeliveryMode: m.DeliveryMode,
				Timestamp:    time.Now(),
				ContentType:  "text/plain",
				Body:         m.Body,
			})

		case rtn := <-notifyReturnChan:
			if rtn.ReplyCode != 0 {
				logger.Error("watchRabbit", rtn, "rabbit msg been return")
			}

		case <-reconnect.C:
			if !brokePipLine {
				break
			}

			logger.Info("watchRabbit", nil, "[RabbitC] disconnect from server, try reconnect...")

			var err error
			if err = r.initRabbit(); err != nil {
				logger.Error("watchRabbit", nil, "[RabbitC] reconnect fail, initRabbit error: "+err.Error())
				break
			}

			if err = r.tryConsume(); err != nil {
				logger.Error("watchRabbit", nil, "[RabbitC] reconnect fail, tryConsume error: "+err.Error())
				_ = r.conn.Close()
				break
			}

			notifyCloseChan = make(chan *amqp.Error)
			notifyCloseChan = r.conn.NotifyClose(notifyCloseChan)

			notifyReturnChan := make(chan amqp.Return)
			notifyReturnChan = r.ch.NotifyReturn(notifyReturnChan)

			brokePipLine = false

			logger.Info("watchRabbit", nil, "[RabbitC] reconnect server success")

		case closeError, ok := <-notifyCloseChan:
			if ok {
				logger.Error("watchRabbit", nil, "[RabbitC] receive close error: "+closeError.Error())
				brokePipLine = true
				r.Wait()
			}
		}
	}

FINISH:
	if !brokePipLine {
		_ = r.conn.Close()
	}
	r.exited <- true
}

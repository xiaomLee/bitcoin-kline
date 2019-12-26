package common

import (
	"bitcoin-kline/logger"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type RabbitC struct {
	url  string
	conn *amqp.Connection
	ch   *amqp.Channel

	asyncPushChan chan Message // push msg chan

	exchangeKeys []string
	queues       map[string]string // queueName: exchange:::router
	consumers    map[string]Consumer

	exit   chan bool
	exited chan bool

	sync.RWMutex
	sync.WaitGroup
}

type Consumer struct {
	name     string
	queue    string
	callback func(amqp.Delivery)
}

var rabbitC *RabbitC

// default rabbit instance
// url is format like 'amqp://account:password@ip:port/vHost'
// for example: amqp://dev:dev@127.0.0.1:5672/dev
func InitRabbit(url string) error {
	rabbitC = NewRabbitC(url)
	return rabbitC.Start()
}

func GetRabbitInstance() *RabbitC {
	return rabbitC
}

func NewRabbitC(url string) *RabbitC {
	return &RabbitC{
		url:           url,
		asyncPushChan: make(chan Message, 1024),
		exchangeKeys:  make([]string, 0),
		queues:        make(map[string]string),
		consumers:     make(map[string]Consumer),
		exit:          make(chan bool),
		exited:        make(chan bool),
	}
}

// Start RabbitC
func (r *RabbitC) Start() (err error) {
	if r.conn, err = amqp.Dial(r.url); err != nil {
		return
	}

	if r.ch, err = r.conn.Channel(); err != nil {
		_ = r.conn.Close()
		return
	}

	go r.watchRabbit()

	return
}

func (r *RabbitC) Restart() (err error) {
	if r.conn, err = amqp.Dial(r.url); err != nil {
		return
	}

	if r.ch, err = r.conn.Channel(); err != nil {
		_ = r.conn.Close()
		return
	}

	for _, exchange := range r.exchangeKeys {
		err = r.ch.ExchangeDeclare(exchange, amqp.ExchangeDirect, true, false, false, false, nil)
		if err != nil {
			return
		}
	}

	for queue, val := range r.queues {
		s := strings.Split(val, ":::")
		if _, err = r.QueueDeclare(queue, s[0], s[1]); err != nil {
			return
		}
	}

	for _, c := range r.consumers {
		if err = r.Consume(c.name, c.queue, c.callback); err != nil {
			return
		}
	}

	return
}

// Close close RabbitC
func (r *RabbitC) Close() {
	close(r.exit)
	r.Wait()
}

func (r *RabbitC) ExchangeDeclare(exchange string) error {
	r.exchangeKeys = append(r.exchangeKeys, exchange)
	return r.ch.ExchangeDeclare(exchange, amqp.ExchangeDirect, true, false, false, false, nil)
}

// when queueName is "", server will create an tmp queue with unique name
func (r *RabbitC) QueueDeclare(queueName, exchange, routerKey string) (string, error) {
	if queue, err := r.ch.QueueDeclare(queueName, false, false, true, false, nil); err != nil {
		return "", err
	} else {
		queueName = queue.Name
		if err = r.ch.QueueBind(queue.Name, routerKey, exchange, false, nil); err != nil {
			return "", err
		}
	}

	r.queues[queueName] = exchange + ":::" + routerKey

	return queueName, nil
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

func (r *RabbitC) Consume(consumer, queue string, c func(amqp.Delivery)) error {
	if consumer == "" {
		return errors.New("consumer can not be null")
	}
	dch, err := r.ch.Consume(queue, consumer, true, false, false, false, nil)
	if err != nil {
		return err
	}

	r.Lock()
	r.consumers[consumer] = Consumer{
		name:     consumer,
		queue:    queue,
		callback: c,
	}
	r.Unlock()

	go r.consumePump(dch)

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

			r.RLock()
			c, ok := r.consumers[delivery.ConsumerTag]
			if !ok {
				break
			}
			r.RUnlock()

			c.callback(delivery)
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
			if err = r.Restart(); err != nil {
				logger.Error("watchRabbit", nil, "[RabbitC] restart fail, initRabbit error: "+err.Error())
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

package provider

import (
	"bitcoin-kline/model"
)

type Provider interface {
	ReadChan(coinType string) <-chan *model.Kline
	StartCollect()
	Stop()
}

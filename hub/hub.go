package hub

import (
	"bitcoin-kline/hub/worker"
)

type Hub struct {
	providerW *worker.ProviderWorker
	klineW    *worker.KlineWorker
	mqW       *worker.MqWorker
	dbW       *worker.DbWorker
}

func NewHub() *Hub {
	return &Hub{
		providerW: worker.NewProviderWorker(),
		klineW:    worker.NewKlineWorker(),
		mqW:       worker.NewMqWorker(),
		dbW:       worker.NewDbWorker(),
	}
}

func (h *Hub) Start() error {
	if err := h.providerW.Start(); err != nil {
		return err
	}
	if err := h.klineW.Start(); err != nil {
		return err
	}
	if err := h.mqW.Start(); err != nil {
		return err
	}
	if err := h.dbW.Start(); err != nil {
		return err
	}

	return nil
}

func (h *Hub) Stop() error {
	h.providerW.Stop()
	h.klineW.Stop()
	h.mqW.Stop()
	h.dbW.Stop()

	return nil
}

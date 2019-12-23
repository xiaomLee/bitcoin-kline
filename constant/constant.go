package constant

const (
	CoinTypeETHUSDT = "ETH/USDT"
	CoinTypeETHCNY  = "ETH/CNY"
	CoinTypeBTCUSDT = "BTC/USDT"
	CoinTypeBTCCNY  = "BTC/CNY"
	CoinTypeRUUSDT  = "RU/USDT"
	CoinTypeRUCNY   = "RU/CNY"
)

const (
	ProviderMock    = "mock"
	ProviderZB      = "zb"
	ProviderHuoBi   = "huobi"
	ProviderOkex    = "okex"
	ProviderBitz    = "bitz"
	ProviderGateio  = "gateio"
	ProviderBinance = "binance"
	ProviderBitmax  = "bitmax"
	ProviderSina    = "sina"
)

const (
	ProviderMockOriginType = 100

	ProviderZBOriginType = iota + 2
	ProviderHuoBiOriginType
	ProviderOkexOriginType
	ProviderBitzOriginType
	ProviderGateioOriginType
	ProviderBinanceOriginType
	ProviderBitmaxOriginType
	ProviderSinaOriginType
)

var ProviderOriginMap = map[int]string{
	ProviderMockOriginType:    ProviderMock,
	ProviderZBOriginType:      ProviderZB,
	ProviderHuoBiOriginType:   ProviderHuoBi,
	ProviderOkexOriginType:    ProviderOkex,
	ProviderBitzOriginType:    ProviderBitz,
	ProviderGateioOriginType:  ProviderGateio,
	ProviderBinanceOriginType: ProviderBinance,
	ProviderBitmaxOriginType:  ProviderBitmax,
	ProviderSinaOriginType:    ProviderSina,
}

const (
	ProviderDataExpireTime = 3 // 供应商数据丢弃时间 秒
)

// mq事件
const (
	MqEventTypeTick = "tick_"
)

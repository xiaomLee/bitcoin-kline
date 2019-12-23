package constant

const (
	CoinTypeETH     = "ETH"
	CoinTypeETHUSDT = "ETH/USDT"
	CoinTypeETHCNY  = "ETH/CNY"
	CoinTypeBTC     = "BTC"
	CoinTypeBTCUSDT = "BTC/USDT"
	CoinTypeBTCCNY  = "BTC/CNY"
	CoinTypeRUUSDT  = "RU/USDT"
	CoinTypeRUCNY   = "RU/CNY"
)

const (
	ProviderMock    = "mock"
	ProviderLocal   = "local"
	ProviderZB      = "zb"
	ProviderHuoBi   = "huobi"
	ProviderBlockcc = "blockcc"
	ProviderOkex    = "okex"
	ProviderBitz    = "bitz"
	ProviderOtcbtc  = "otcbtc"
	ProviderGateio  = "gateio"
	ProviderBinance = "binance"
	ProviderBitmax  = "bitmax"
	ProviderSina    = "sina"
)

const (
	ProviderMockOriginType = 100

	ProviderLocalOriginType = iota + 2
	ProviderZBOriginType
	ProviderHuoBiOriginType
	ProviderBlockccOriginType
	ProviderOkexOriginType
	ProviderBitzOriginType
	ProviderOtcbtcOriginType
	ProviderGateioOriginType
	ProviderBinanceOriginType
	ProviderBitmaxOriginType
	ProviderSinaOriginType
)

var ProviderOriginMap = map[int]string{
	ProviderMockOriginType:    ProviderMock,
	ProviderLocalOriginType:   ProviderLocal,
	ProviderZBOriginType:      ProviderZB,
	ProviderHuoBiOriginType:   ProviderHuoBi,
	ProviderBlockccOriginType: ProviderBlockcc,
	ProviderOkexOriginType:    ProviderOkex,
	ProviderBitzOriginType:    ProviderBitz,
	ProviderOtcbtcOriginType:  ProviderOtcbtc,
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
	MqEventTypeTicker       = "ticker_"
	MqEventTypeKline        = "kline_"
	MqEventTypeExchangeRate = "exchange_rate"
)

const (
	CurrencyCny  = "CNY"
	CurrencyUSDT = "USDT"
	CurrencyTEST = "TEST"
)

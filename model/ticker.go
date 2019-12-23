package model

// 成交行情
type Ticker struct {
	CoinType   string
	Price      string `json:"price"`      // 成交价
	Vol        int    `json:"vol"`        // 成交数量
	TickerType int    `json:"tickerType"` // 涨跌 -1=跌 1=涨
	CreateTime int64  `json:"createTime"`
}

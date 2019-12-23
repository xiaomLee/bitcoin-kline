package model

import (
	"bitcoin-kline/common"
	"bitcoin-kline/config"
)

type Kline struct {
	Id          int64  `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"-"` // id
	CoinType    string `gorm:"column:coinType" json:"coinType"`               // 币种
	High        string `gorm:"column:high" json:"high"`                       // 最高报价
	Low         string `gorm:"column:low" json:"low"`                         // 最低报价
	Open        string `gorm:"column:open" json:"open"`                       // 开盘价
	Close       string `gorm:"column:close" json:"close"`                     // 收盘价
	CreateTime  int64  `gorm:"column:createTime" json:"time"`                 // 时间
	UpdateTime  int64  `gorm:"column:updateTime" json:"-"`                    // 最后更新时间
	TimeScale   string `gorm:"column:timeScale" json:"-"`                     // 分时图刻度
	Origin      int    `gorm:"column:origin" json:"origin"`                   // 是否原始数据：1:是，0：否
	OriginPrice string `gorm:"column:originPrice" json:"-"`                   // 原始报价 市场价
	Volume      string `gorm:"column:volume" json:"volume"`                   // 24小时成交量

	// 标识本地风控数据
	RiskType  int `gorm:"column:-" json:"-"` // 风控类型 0=未风控 1=开始调整 2=恢复调整
	TotalStep int `gorm:"column:-" json:"-"` // 总步数 = 调整时间 或 恢复时间
	Step      int `gorm:"column:-" json:"-"` // 当前步数
}

func (k *Kline) TableName() string {
	return "kline"
}

func (k *Kline) Copy() Kline {
	return Kline{
		Id:          k.Id,
		CoinType:    k.CoinType,
		High:        k.High,
		Low:         k.Low,
		Open:        k.Open,
		Close:       k.Close,
		CreateTime:  k.CreateTime,
		UpdateTime:  k.UpdateTime,
		TimeScale:   k.TimeScale,
		Origin:      k.Origin,
		OriginPrice: k.OriginPrice,
		Volume:      k.Volume,
		RiskType:    k.RiskType,
		TotalStep:   k.TotalStep,
		Step:        k.Step,
	}
}

func GetDbKline(coinType string, createTime int64, timeScale string) *Kline {
	val, ok := config.TimeScaleMap[timeScale]
	if !ok {
		return nil
	}
	createTime = createTime - createTime%int64(val)

	var item Kline
	db := common.MustGetDB("futures")
	if err := db.Where("coinType=? and TimeScale=? and createTime=?", coinType, timeScale, createTime).Find(&item).Error; err != nil {
		return nil
	}

	return &item
}

// 正序
func GetKlineHistoryClose(coinType string, timeScale string, endTime int64, size int) []float64 {
	ret := make([]float64, 0)

	db := common.MustGetDB("futures")
	var list []struct {
		Price float64 `gorm:"column:price"`
	}
	db = db.Table("kline").Select("cast(close as decimal(20,4)) as price")
	db = db.Where("coinType=? and timeScale=? and createTime<?", coinType, timeScale, endTime).Order("createTime desc").Limit(size)
	if err := db.Find(&list).Error; err != nil {
		return ret
	}

	for i := len(list) - 1; i >= 0; i-- {
		ret = append(ret, list[i].Price)
	}
	return ret
}

package worker

import (
	"bitcoin-kline/common"
	"bitcoin-kline/config"
	"bitcoin-kline/logger"
	"bitcoin-kline/model"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type DbWorker struct {
	breakMainLogic chan bool // 结束命令管道
	exited         chan bool // 确认结束命令管道
}

const DefaultDbChanSize = 1024

var (
	klineDbChan chan model.Kline
	klineCache  []model.Kline
)

func NewDbWorker() *DbWorker {
	klineDbChan = make(chan model.Kline, DefaultDbChanSize)
	klineCache = make([]model.Kline, 0)

	return &DbWorker{
		breakMainLogic: make(chan bool),
		exited:         make(chan bool),
	}
}

func (w *DbWorker) Start() error {
	go w.workLoop()
	return nil
}

// 结束主逻辑
func (w *DbWorker) Stop() {
	close(w.breakMainLogic)
	<-w.exited
}

func (w *DbWorker) workLoop() {
	timer := time.NewTimer(time.Minute * 1)

	for {
		select {
		case kline := <-klineDbChan:
			// 随机生成成交量
			kline.Volume = strconv.Itoa(rand.Intn(100))

			// 保存秒级数据
			if err := saveTick(kline); err != nil {
				logger.Error("DbWorker_saveTicker", err, "DbWorker saveTick err")
			}

			// 构造分时数据并保存
			scaleItems := make([]model.Kline, 0)
			for scale, val := range config.TimeScaleMap {
				item := kline.Copy()
				item.TimeScale = scale
				item.CreateTime = item.CreateTime - item.CreateTime%int64(val)

				scaleItems = append(scaleItems, item)
			}
			if err := saveKline2DB(scaleItems); err != nil {
				logger.Error("DbWorker_saveKline2DB", err, "DbWorker saveKline2DB err")
			}

		case <-timer.C:
			if err := deleteOldTick(); err != nil {
				logger.Error("DbWorker_deleteOldTicker", err, "DbWorker deleteOldTick err")
			}
			// todo
			timer.Reset(time.Hour * 6)

		case <-w.breakMainLogic:
			if err := flushTickCache2DB(); err != nil {
				logger.Error("DbWorker_flushTicker", err, "DbWorker flushTicker to db err")
			}
			goto EXIT
		}
	}
EXIT:
	w.exited <- true
}

// 批量将kline数据写入或更新到数据库
func saveKline2DB(items []model.Kline) error {
	db := common.MustGetDB("kline")

	sql := "insert into kline (coinType, high, low, open, close, createTime, updateTime, timeScale, origin, originPrice, volume) values "
	values := []string{}
	for _, item := range items {
		values = append(values, fmt.Sprintf("('%s', '%s', '%s', '%s', '%s', %v, %v, '%s', %d, '%s', '%s')",
			item.CoinType, item.High, item.Low, item.Open, item.Close, item.CreateTime, item.UpdateTime, item.TimeScale, item.Origin, item.OriginPrice, item.Volume))
	}
	sql += strings.Join(values, ",")
	sql += " ON DUPLICATE KEY UPDATE " +
		"low=if(cast(low as decimal(20,4)) < cast(values(low) as decimal(20,4)), low, values(low)), " +
		"high=if(cast(high as decimal(20,4)) > cast(values(high) as decimal(20,4)), high, values(high)), " +
		"volume=cast(volume as decimal(20,4)) + cast(values(volume) as decimal(20,4)), " +
		"close=values(close), updateTime=values(updateTime)"

	err := db.Exec(sql).Error
	if err != nil {
		logger.Error("DbWorker_saveKline2DB", sql, "saveKline2DB sql err")
	}
	return err
}

func saveTick(item model.Kline) error {
	klineCache = append(klineCache, item)
	// 缓存大于50 执行插入db
	if len(klineCache) < 60 {
		return nil
	}

	return flushTickCache2DB()
}

func flushTickCache2DB() error {
	if len(klineCache) == 0 {
		return nil
	}
	db := common.MustGetDB("kline")

	sql := "insert into tick_cache (coinType, high, low, open, close, createTime, updateTime, timeScale, origin, originPrice, volume) values "
	values := []string{}
	for _, item := range klineCache {
		values = append(values, fmt.Sprintf("('%s', '%s', '%s', '%s', '%s', %v, %v, '%s', %d, '%s', '%s')",
			item.CoinType, item.High, item.Low, item.Open, item.Close, item.CreateTime, item.UpdateTime, item.TimeScale, item.Origin, item.OriginPrice, item.Volume))
	}
	sql += strings.Join(values, ",")
	sql += " ON DUPLICATE KEY UPDATE open=values(open), close=values(close), high=values(high), low=values(low), updateTime=values(updateTime), volume=values(volume)"
	err := db.Exec(sql).Error
	if err != nil {
		logger.Error("DbWorker_saveTicker2DB", sql, "saveTick sql err")
		return err
	}
	//清空缓存
	klineCache = klineCache[:0]
	return nil
}

func deleteOldTick() error {
	// 删除6小时前的数据
	db := common.MustGetDB("kline")
	sql := fmt.Sprintf("delete from tick_cache where createTime < %v", time.Now().Unix()-60*10)
	err := db.Exec(sql).Error
	if err != nil {
		logger.Error("DbWorker_saveTicker2DB", sql, "saveTick sql err")
	}

	return err
}

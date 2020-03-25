package main

import (
	"bitcoin-kline/common"
	"bitcoin-kline/config"
	"bitcoin-kline/hub"
	"bitcoin-kline/logger"
	"bitcoin-kline/router"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/judwhite/go-svc/svc"
)

type BaseServer struct {
	server *http.Server
	worker *hub.Hub
}

func (s *BaseServer) Init(env svc.Environment) error {
	config.InitConfig()
	println("RunMode:", config.CURMODE)
	for key, val := range config.GetSection("system") {
		println(key, val)
	}

	// init log
	logger.ConfigLogger(
		config.CURMODE,
		config.GetConfig("system", "app_name"),
		config.GetConfig("logs", "dir"),
		config.GetConfig("logs", "file_name"),
		config.GetConfigInt("logs", "keep_days"),
		config.GetConfigInt("logs", "rate_hours"),
	)
	println("logger init success")

	// init mysql
	dbInfo := config.GetSection("dbInfo")
	for name, info := range dbInfo {
		if err := common.AddDB(
			name,
			info,
			config.GetConfigInt("mysql", "maxConn"),
			config.GetConfigInt("mysql", "idleConn"),
			time.Hour*time.Duration(config.GetConfigInt("mysql", "maxLeftTime"))); err != nil {
			return err
		}
	}
	println("mysql init success")

	// init redis
	//if err := common.AddRedisInstance(
	//	"",
	//	config.GetConfig("redis", "addr"),
	//	config.GetConfig("redis", "port"),
	//	config.GetConfig("redis", "password"),
	//	config.GetConfigInt("redis", "db_num")); err != nil {
	//	return err
	//}
	//println("redis init success")

	rabbitUrl := fmt.Sprintf("amqp://%s:%s@%s:%s%s",
		config.GetConfig("rabbit", "account"),
		config.GetConfig("rabbit", "password"),
		config.GetConfig("rabbit", "ip"),
		config.GetConfig("rabbit", "port"),
		config.GetConfig("rabbit", "vhost"),
	)
	if err := common.InitRabbit(rabbitUrl); err != nil {
		return err
	}
	println("rabbit init success")

	return nil
}

func (s *BaseServer) Start() error {
	s.server = &http.Server{
		Addr:    ":" + config.GetConfig("system", "http_listen_port"),
		Handler: router.NewEngine(),
	}
	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				panic(err)
			}
		}
	}()
	println("http service start")

	s.worker = hub.NewHub()
	if err := s.worker.Start(); err != nil {
		panic(err)
	}
	println("booster service start")

	return nil
}

func (s *BaseServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		println(err.Error())
	}
	println("http server shutdown")

	if err := s.worker.Stop(); err != nil {
		println(err.Error())
	}
	println("booster service stop")

	// release source
	common.ReleaseMysqlDBPool()
	common.ReleaseRedisPool()
	common.CloseRabbit()
	println("release source success")

	return nil
}

func main() {
	if err := svc.Run(&BaseServer{}); err != nil {
		println(err.Error())
	}
}

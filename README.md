# btcoin-kline
聚合数字货币实时行情，并推送至rabbitMQ。实现外部用户订阅（待实现）



## 依赖环境
    本项目使用go mod来管理依赖，原则上go版本大于1.11即可。本人环境为go1.13
    golang >= 1.11
    mysql、rabbitmq
    
    
## docker一键安装依赖环境
```bash
docker-compose -f ./kline-docker.yaml
```
    
## 修改配置文件
    系统默认为dev开发环境，可通过设置环境变量RUNMODE来修改，支持dev、test、online
    Windows：系统->环境变量
    Macos：export RUNMODE=test
    Linux：export RUNMODE=test
    修改./conf/下面相应的配置项
 
## 数据库初始化
    执行kline.sql里的语句
   
## 编译运行
    go build -o kline main.go
    ./kline
    
## 项目结构
    ├── common              // 公共库
    ├── conf
    ├── config
    ├── constant
    ├── hub                 // 数据采集核心代码
    │   ├── provider        // 数据采集器 大部分服务需翻墙访问
    │   │   ├── binance
    │   │   ├── bitmax
    │   │   ├── bitz
    │   │   ├── gateio
    │   │   ├── huobi
    │   │   ├── mock        // 开发mock数据用
    │   │   ├── okex
    │   │   ├── sina
    │   │   └── zb
    │   └── worker          // 具体的任务worker，协作完成整个数据采集任务
    ├── logger              // 日志库
    ├── logs
    ├── middleware
    ├── model
    └── router              // http路由，后续订阅服务使用
    
## 开发tips
    1. 若不想使用rabbitMq的消息服务,可在hub/hob.go里面注释掉MQ相关的worker.同时还可在main.go里注释rabbitMq的启动init
    2. provider目录下有个provider_test.go的单例测试,修改相应代码可测试每个provider的数据
    3. 目前该项目支持采集的币种配置在config/config.go中,查看SupportCoinTypes
    4. dev环境仅支持mock数据,可在hub/worker/providerworker.go的50行进行修改
    5. 每个provider采集器可添加代理实现翻墙,具体代码可参照zb采集的第142行.后续有空会将添加代理的功能抽成配置
    
    
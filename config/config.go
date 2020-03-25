package config

import (
	"bitcoin-kline/constant"
	"fmt"
	"os"
	"strconv"

	"github.com/widuu/goini"
)

const (
	ENV_DEV    = "dev"
	ENV_TEST   = "test"
	ENV_ONLINE = "online"
	ENV_PRE    = "pre"

	ModeEnv = "RUNMODE"
)

var (
	CURMODE = ""

	SupportCoinTypes = []string{
		constant.CoinTypeETHUSDT,
		constant.CoinTypeBTCUSDT,
		constant.CoinTypeRUCNY,
	}
	TimeScaleMap = map[string]int{
		"1":  60,
		"5":  60 * 5,
		"15": 60 * 15,
		"60": 60 * 60,
		"1D": 3600 * 24,
		"7D": 86400 * 7,
		"1M": 86400 * 30,
	}

	configData []map[string]map[string]string
)

func InitConfig() {
	mode := os.Getenv(ModeEnv)
	if mode == "" {
		// TODO online should set RUNMODE Env
		//panic("env " + ModeEnv + " not set")
		mode = "dev"
	}
	if mode != ENV_DEV && mode != ENV_TEST && mode != ENV_ONLINE && mode != ENV_PRE {
		panic("env " + ModeEnv + " should be: dev, test, online, and pre")
	}
	CURMODE = mode

	var fileName = fmt.Sprintf("./conf/%s.ini", CURMODE)
	if _, err := os.Stat(fileName); err != nil {
		if os.IsNotExist(err) {
			panic("configuration file " + fileName + " is not exist")
		}
		panic("configuration file " + fileName + " is privilege mode is not right")
	}

	conf := goini.SetConfig(fileName)
	configData = conf.ReadList()
}

func GetConfig(section string, key string) string {
	for _, v := range configData {
		if _, ok := v[section]; ok {
			return v[section][key]
		}
	}
	return ""
}

func GetConfigInt(section string, key string) int {
	v := GetConfig(section, key)
	if v == "" {
		return 0
	}
	n, _ := strconv.Atoi(v)
	return n
}

func GetConfigInt64(section string, key string) int64 {
	v := GetConfig(section, key)
	if v == "" {
		return 0
	}
	n, _ := strconv.ParseInt(v, 10, 64)
	return n
}

func GetConfigFloat64(section string, key string) float64 {
	v := GetConfig(section, key)
	if v == "" {
		return 0
	}
	n, _ := strconv.ParseFloat(v, 64)
	return n
}

func GetSection(section string) map[string]string {
	for _, v := range configData {
		if _, ok := v[section]; ok {
			return v[section]
		}
	}
	return nil
}

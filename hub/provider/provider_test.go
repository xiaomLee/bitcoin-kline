package provider

import (
	"bitcoin-kline/hub/provider/binance"
	"bitcoin-kline/hub/provider/bitmax"
	"bitcoin-kline/hub/provider/bitz"
	"bitcoin-kline/hub/provider/gateio"
	"bitcoin-kline/hub/provider/huobi"
	"bitcoin-kline/hub/provider/mock"
	"bitcoin-kline/hub/provider/okex"
	"bitcoin-kline/hub/provider/sina"
	"bitcoin-kline/hub/provider/zb"
	"fmt"
	"math"
	"sort"
	"testing"

	"github.com/smallnest/weighted"
)

func TestProvider(t *testing.T) {
	coinType1 := "RU/CNY"
	//coinType2 := "RU/USDT"
	p := newProvider("mock")
	p.StartCollect()
	c1 := p.ReadChan(coinType1)
	//c2 := p.ReadChan(coinType2)
	go func() {
		for {
			select {
			case item := <-c1:
				fmt.Printf("%s: %+v \n", coinType1, item)
				//case <-c2:
				//fmt.Printf("%s: %+v \n", coinType2, item)
			}
		}
	}()
	//time.Sleep(20 * time.Second)
	select {}
	t.Log("stop server")
	p.Stop()
	t.Log("stop success")
}

func newProvider(name string) Provider {
	switch name {
	case "mock":
		return mock.NewProvider()
	case "zb":
		return zb.NewZbProvider()
	case "huobi":
		return huobi.NewProvider()
	case "okex":
		return okex.NewProvider()
	case "bitz":
		return bitz.NewProvider()
	case "gateio":
		return gateio.NewProvider()
	case "binance":
		return binance.NewProvider()
	case "bitmax":
		return bitmax.NewProvider()
	case "sina":
		return sina.NewProvider()

	}
	return mock.NewProvider()
}

func TestXXX(t *testing.T) {
	p1 := []float64{10, 3, 4.6, 5, 6.5, 1, 7.0, 8, 15, 20}
	p2 := []float64{7549.29, 7550.7, 7551.1600, 7552.22, 7553.1800, 7554.94}

	filterOutliers(p1)
	filterOutliers(p2)
}

func filterOutliers(prices []float64) {
	sort.Slice(prices, func(i, j int) bool {
		return prices[i] < prices[j]
	})

	length := len(prices)

	q1 := 0.25*prices[int(math.Floor(float64((length+1)/4)))] + 0.75*prices[int(math.Ceil(float64((length+1)/4)))]
	q3 := 0.25*prices[int(math.Floor(float64((length+1)*3/4)))] + 0.75*prices[int(math.Ceil(float64((length+1)*3/4)))]
	iqr := q3 - q1
	println("q1", q1, "q3", q3, "iqr", iqr)

	max := q3 + iqr*1.5
	min := q1 - iqr*1.5
	result := make([]float64, 0)
	for _, price := range prices {
		if price <= min || price >= max {
			continue
		}
		result = append(result, price)
	}
	fmt.Println(result)
}

func TestWeight_x(t *testing.T) {
	w := weighted.NewRandW()
	w.Add("server1", 5)
	w.Add("server2", 2)
	w.Add("server3", 3)

	results := make(map[string]int)

	for i := 0; i < 100; i++ {
		s := w.Next().(string)
		results[s]++
	}
	fmt.Printf("%+v \n", results)

	if results["server1"] != 50 || results["server2"] != 20 || results["server3"] != 30 {
		t.Error("the algorithm is wrong")
	}

	w.Reset()
	results = make(map[string]int)

	for i := 0; i < 100; i++ {
		s := w.Next().(string)
		results[s]++
	}
	fmt.Printf("%+v \n", results)

	if results["server1"] != 50 || results["server2"] != 20 || results["server3"] != 30 {
		t.Error("the algorithm is wrong")
	}

	w.RemoveAll()
	w.Add("server1", 7)
	w.Add("server2", 9)
	w.Add("server3", 13)

	results = make(map[string]int)

	for i := 0; i < 29000; i++ {
		s := w.Next().(string)
		results[s]++
	}
	fmt.Printf("%+v \n", results)
	if results["server1"] != 7000 || results["server2"] != 9000 || results["server3"] != 13000 {
		t.Error("the algorithm is wrong")
	}
}

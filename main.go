package main

import (
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/gateio/gateapi-go/v6"
)

var logger = log.New(flag.CommandLine.Output(), "", log.LstdFlags)

func panicGateError(err error) {
	if e, ok := err.(gateapi.GateAPIError); ok {
		logrus.Fatal(fmt.Sprintf("Gate API error, label: %s, message: %s", e.Label, e.Message))
	}
	log.Fatal(err)
}

func main() {
	var baseUrl string
	var key, secret, orderAmount, orderPrice, currencyPair, functionName string
	flag.StringVar(&key, "k", "", "Gate APIv4 key")
	flag.StringVar(&secret, "s", "", "Gate APIv4 secret")
	flag.StringVar(&baseUrl, "u", "", "API based URL used")
	flag.StringVar(&orderAmount, "a", "", "order amount")
	flag.StringVar(&orderPrice, "p", "", "order price")
	flag.StringVar(&currencyPair, "cp", "", "currency pair")
	flag.StringVar(&functionName, "func", "check", "function selection")
	flag.Parse()
	usage := fmt.Sprintf("Usage: %s -func <buy/sell/cancel/check> -k <api-key> -s <api-secret> -a <order-amount> -p <order-price> -cp <currency-pair>", os.Args[0])
	baseDir := filepath.Dir(os.Args[0])
	err := InitLogger(filepath.Join(baseDir, "/log/spot.log"), 30)
	if err != nil {
		panic(err)
	}
	if key == "" || secret == "" {
		logger.Println(key)
		logger.Println(secret)
		logger.Println(usage)
		flag.PrintDefaults()
		os.Exit(1)
	}
	//f, _ := strconv.ParseFloat(orderAmount, 64)
	/*	if f <= 0 {
		logger.Println("order amount should bigger than 0")
		flag.PrintDefaults()
		os.Exit(1)
	}*/

	/*	if flag.NArg() < 0 {
		logger.Println(flag.NArg())
		logger.Println(usage)
		flag.PrintDefaults()
		os.Exit(1)
	}*/
	runConfig, err := NewRunConfig(key, secret, &baseUrl)
	if err != nil {
		logrus.WithFields(logrus.Fields{"config setting": runConfig}).Fatal("当前配置出错")
		os.Exit(1)
	}
	logrus.Info("launch auto spot services")
	logrus.WithFields(logrus.Fields{"config setting": runConfig}).Info("当前配置")
	rand.Seed(time.Now().Unix())
	currencyPair2, _ := GetCurrentPair(currencyPair)
	fmt.Println(currencyPair2)

	switch functionName {
	case "buy":
		logrus.Info("start spot buy")
		SpotBuy(runConfig, orderAmount, orderPrice, currencyPair)
	case "check":
		logrus.Info("start list tickers")
		ListTickers(runConfig, currencyPair)
	case "sell":
		logrus.Info("start sell spot")
	case "cancel":
		logrus.Info("cancel spot")
	default:
		logrus.Fatal("Invalid function provided. Available: spot, check")
	}

}

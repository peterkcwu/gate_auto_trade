package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/gateio/gateapi-go/v6"
)

var logger = log.New(flag.CommandLine.Output(), "", log.LstdFlags)

func panicGateError(err error) {
	if e, ok := err.(gateapi.GateAPIError); ok {
		log.Fatal(fmt.Sprintf("Gate API error, label: %s, message: %s", e.Label, e.Message))
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
	usage := fmt.Sprintf("Usage: %s -func <spot/check> -k <api-key> -s <api-secret> -a <order-amount> -p <order-price> -cp <currency-pair>", os.Args[0])
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
	logger.Println(flag.NArg())
	/*	if flag.NArg() < 0 {
		logger.Println(flag.NArg())
		logger.Println(usage)
		flag.PrintDefaults()
		os.Exit(1)
	}*/
	runConfig, err := NewRunConfig(key, secret, &baseUrl)
	if err != nil {
		logger.Fatal(err)
	}
	fmt.Println(runConfig)
	rand.Seed(time.Now().Unix())
	currencyPair2, _ := GetCurrentPair(currencyPair)
	fmt.Println(currencyPair2)
	args := flag.Args()
	fmt.Println(args)

	switch functionName {
	case "spot":
		SpotBuy(runConfig, orderAmount, orderPrice)
	case "check":
		ListTickers(runConfig, currencyPair)
	default:
		logger.Fatal("Invalid function provided. Available: spot, check")
	}

}

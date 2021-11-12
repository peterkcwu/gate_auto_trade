package main

import (
	"context"
	"fmt"
	"github.com/antihax/optional"
	"github.com/gateio/gateapi-go/v5"
	"github.com/shopspring/decimal"
)

func GetCurrentPair(currencyPair string) (gateapi.CurrencyPair, error) {
	client := gateapi.NewAPIClient(gateapi.NewConfiguration())
	// uncomment the next line if your are testing against testnet
	// client.ChangeBasePath("https://fx-api-testnet.gateio.ws/api/v4")
	ctx := context.Background()
	//currencyPair := "ETH_BTC" // string - Currency pair

	result, _, err := client.SpotApi.GetCurrencyPair(ctx, currencyPair)
	if err != nil {
		return gateapi.CurrencyPair{}, err
	} else {
		return result, nil
	}
}

func ListTickers(config *RunConfig, currencyPair string) {
	client := gateapi.NewAPIClient(gateapi.NewConfiguration())
	// Setting host is optional. It defaults to https://api.gateio.ws/api/v4
	client.ChangeBasePath(config.BaseUrl)
	ctx := context.WithValue(context.Background(), gateapi.ContextGateAPIV4, gateapi.GateAPIV4{
		Key:    config.ApiKey,
		Secret: config.ApiSecret,
	})

	cp, _, err := client.SpotApi.GetCurrencyPair(ctx, currencyPair)
	if err != nil {
		panicGateError(err)
	}

	tickers, _, err := client.SpotApi.ListTickers(ctx, &gateapi.ListTickersOpts{CurrencyPair: optional.NewString(cp.Id)})
	if err != nil {
		panicGateError(err)
	}
	logger.Println(tickers)
}

func SpotBuy(config *RunConfig, orderAmount string, orderPrice string) {
	client := gateapi.NewAPIClient(gateapi.NewConfiguration())
	// Setting host is optional. It defaults to https://api.gateio.ws/api/v4
	client.ChangeBasePath(config.BaseUrl)
	ctx := context.WithValue(context.Background(), gateapi.ContextGateAPIV4, gateapi.GateAPIV4{
		Key:    config.ApiKey,
		Secret: config.ApiSecret,
	})

	currencyPair := "GT_USDT"
	currency := "USDT"
	cp, _, err := client.SpotApi.GetCurrencyPair(ctx, currencyPair)
	if err != nil {
		panicGateError(err)
	}
	logger.Printf("testing against currency pair: %s\n", cp.Id)
	//以usdt为最小单位
	minAmount := cp.MinQuoteAmount

	tickers, _, err := client.SpotApi.ListTickers(ctx, &gateapi.ListTickersOpts{CurrencyPair: optional.NewString(cp.Id)})
	if err != nil {
		panicGateError(err)
	}
	logger.Println(tickers)
	//用最新成交价
	//lastPrice := tickers[0].Last
	lastPrice := orderPrice

	// better avoid using float, take the following decimal library for example
	// `go get github.com/shopspring/decimal`
	//orderAmount := decimal.RequireFromString(minAmount).Mul(decimal.NewFromInt32(2))

	decimalOrderAmount, err := decimal.NewFromString(orderAmount)
	if err != nil {
		logger.Fatal("decimal newFromString error")
	}
	if decimalOrderAmount.Cmp(decimal.RequireFromString(minAmount)) == -1 {
		logger.Fatal(fmt.Sprintf("order amount less than mini amount, must order bigger than %s", minAmount))
	}
	//以最小数量的倍数购买
	//orderAmountFinal := decimal.RequireFromString(minAmount).Mul(decimalOrderAmount)

	balance, _, err := client.SpotApi.ListSpotAccounts(ctx, &gateapi.ListSpotAccountsOpts{Currency: optional.NewString(currency)})
	if err != nil {
		panicGateError(err)
	}
	if decimal.RequireFromString(balance[0].Available).Cmp(decimalOrderAmount) < 0 {
		logger.Fatal("balance not enough")
	}

	newOrder := gateapi.Order{
		Text:         "t-my-custom-id", // optional custom order ID
		CurrencyPair: cp.Id,
		Type:         "limit",
		Account:      "spot", // create spot order. set to "margin" if creating margin orders
		Side:         "buy",
		Amount:       decimalOrderAmount.String(),
		Price:        lastPrice, // use last price
		TimeInForce:  "gtc",
		AutoBorrow:   false,
	}
	logger.Printf("place a spot %s order in %s with amount %s and price %s\n", newOrder.Side, newOrder.CurrencyPair, newOrder.Amount, newOrder.Price)
	createdOrder, _, err := client.SpotApi.CreateOrder(ctx, newOrder)
	if err != nil {
		panicGateError(err)
	}
	logger.Printf("order created with ID: %s, status: %s\n", createdOrder.Id, createdOrder.Status)
	for createdOrder.Status == "open" {
		order, _, err := client.SpotApi.GetOrder(ctx, createdOrder.Id, createdOrder.CurrencyPair)
		if err != nil {
			panicGateError(err)
		}
		logger.Printf("order %s filled: %s, left: %s\n", order.Id, order.FilledTotal, order.Left)
		if order.Status == "closed" {
			break
		}
	}
	/*	result, _, err := client.SpotApi.CancelOrder(ctx, createdOrder.Id, createdOrder.CurrencyPair)
		if err != nil {
			panicGateError(err)
		}
		if result.Status == "cancelled" {
			logger.Printf("order %s cancelled\n", createdOrder.Id)
		}*/

	// order finished
	trades, _, err := client.SpotApi.ListMyTrades(ctx, createdOrder.CurrencyPair,
		&gateapi.ListMyTradesOpts{OrderId: optional.NewString(createdOrder.Id)})
	if err != nil {
		panicGateError(err)
	}
	for _, t := range trades {
		logger.Printf("order %s filled %s with price: %s\n", t.OrderId, t.Amount, t.Price)
	}

}

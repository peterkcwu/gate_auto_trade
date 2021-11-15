package main

import (
	"context"
	"fmt"
	"github.com/antihax/optional"
	"github.com/gateio/gateapi-go/v5"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
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
		logrus.WithFields(logrus.Fields{
			"Current pair": result,
		}).Info("当前CP")
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
	logrus.WithFields(logrus.Fields{
		"Ticker": tickers[0],
	}).Info("当前USDT ticker")
	logger.Println(tickers)
}

func SpotBuy(config *RunConfig, orderAmount string, orderPrice string, currencyPair string) {

	if i, _ := strconv.Atoi(orderAmount); i > 50 {
		logger.Printf("place a spot amount %s\n， bigger than 50", orderAmount)
		logrus.Warn("oder amount > 50,cannot buy")
		return
	}
	client := gateapi.NewAPIClient(gateapi.NewConfiguration())
	// Setting host is optional. It defaults to https://api.gateio.ws/api/v4
	client.ChangeBasePath(config.BaseUrl)
	ctx := context.WithValue(context.Background(), gateapi.ContextGateAPIV4, gateapi.GateAPIV4{
		Key:    config.ApiKey,
		Secret: config.ApiSecret,
	})

	//用来检查钱当前币种余额
	currency := "USDT"
	cp, _, err := client.SpotApi.GetCurrencyPair(ctx, currencyPair)
	if err != nil {
		panicGateError(err)
	}
	logger.Printf("buyying against currency pair: %s\n", cp.Id)
	//以usdt为最小单位 例如买GT币最少1u,orderAmount最少0.2
	minAmount := cp.MinBaseAmount

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
	if minAmount != "" {
		if decimalOrderAmount.Cmp(decimal.RequireFromString(minAmount)) == -1 {
			logger.Fatal(fmt.Sprintf("order amount less than mini amount, must order bigger than %s", minAmount))
		}
	}
	//以最小数量的倍数购买
	//orderAmountFinal := decimal.RequireFromString(minAmount).Mul(decimalOrderAmount)

	balance, _, err := client.SpotApi.ListSpotAccounts(ctx, &gateapi.ListSpotAccountsOpts{Currency: optional.NewString(currency)})
	if err != nil {
		panicGateError(err)
	}
	orderTotalUSDPrice := decimalOrderAmount.Mul(decimal.RequireFromString(lastPrice))
	if decimal.RequireFromString(balance[0].Available).Cmp(orderTotalUSDPrice) < 0 {
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
	logrus.WithFields(logrus.Fields{"order side": newOrder.Side,
		"order cp": newOrder.CurrencyPair,
		"挂单购买数量":   newOrder.Amount,
		"挂单购买价格":   newOrder.Price,
	}).Info("准备挂单")
	createdOrder, _, err := client.SpotApi.CreateOrder(ctx, newOrder)
	if err != nil {
		panicGateError(err)
	}
	logger.Printf("order created with ID: %s, status: %s\n", createdOrder.Id, createdOrder.Status)
	logrus.WithFields(logrus.Fields{"order ID": createdOrder.Id,
		"order cp":     newOrder.CurrencyPair,
		"order status": createdOrder.Status,
	}).Info("挂单成功")
	for createdOrder.Status == "open" {
		order, _, err := client.SpotApi.GetOrder(ctx, createdOrder.Id, createdOrder.CurrencyPair)
		if err != nil {
			panicGateError(err)
		}
		logrus.WithFields(logrus.Fields{"order id": order.Id,
			"order FilledTotal": order.FilledTotal,
			"order Left":        order.Left,
			"购买数量":              order.Amount,
			"购买价格":              order.Price,
		}).Info("等待挂单完全成交")
		time.Sleep(time.Millisecond * 500)
		if order.Status == "closed" {
			logger.Printf("order %s filled: %s, left: %s\n", order.Id, order.FilledTotal, order.Left)
			logrus.Info(fmt.Sprintf("order %s filled: %s, left: %s\n", order.Id, order.FilledTotal, order.Left))
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
		logrus.WithFields(logrus.Fields{"order id": t.OrderId,
			"成交数量": t.Amount,
			"成交价格": t.Price,
		}).Info("已成交")
	}
	sellPrice := decimal.RequireFromString(orderPrice).Mul(decimal.NewFromInt(2))
	sellPriceStr := sellPrice.String()
	SpotSell(config, orderAmount, sellPriceStr, currencyPair)

}

func SpotSell(config *RunConfig, orderAmount string, orderPrice string, currencyPair string) {

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
	logger.Printf("selling against currency pair: %s\n", cp.Id)
	//以usdt为最小单位 例如买GT币最少1u,orderAmount最少0.2
	minAmount := cp.MinBaseAmount

	tickers, _, err := client.SpotApi.ListTickers(ctx, &gateapi.ListTickersOpts{CurrencyPair: optional.NewString(cp.Id)})
	if err != nil {
		panicGateError(err)
	}
	logger.Println(tickers)
	//用最新成交价
	//lastPrice := tickers[0].Last
	decimalOrderPrice, err := decimal.NewFromString(orderPrice)
	if err != nil {
		logger.Fatal("decimal newFromString error")
	}
	var lastPrice string
	//decimalOrderPrice < tickers[0].Last
	if decimalOrderPrice.Cmp(decimal.RequireFromString(tickers[0].Last)) == -1 {
		logger.Println(fmt.Sprintf("出价低于成交价，使用目前成交价 %s 出价", tickers[0].Last))
		logrus.Info(fmt.Sprintf("出价低于成交价，使用目前成交价 %s 出价", tickers[0].Last))
		lastPrice = tickers[0].Last
	} else {
		lastPrice = orderPrice
	}

	decimalOrderAmount, err := decimal.NewFromString(orderAmount)
	if err != nil {
		logger.Fatal("decimal newFromString error")
	}
	if minAmount != "" {
		if decimalOrderAmount.Cmp(decimal.RequireFromString(minAmount)) == -1 {
			logger.Fatal(fmt.Sprintf("order amount less than mini amount, must order bigger than %s", minAmount))
		}
	}

	newOrder := gateapi.Order{
		Text:         "t-my-custom-id", // optional custom order ID
		CurrencyPair: cp.Id,
		Type:         "limit",
		Account:      "spot", // create spot order. set to "margin" if creating margin orders
		Side:         "sell",
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
	logrus.WithFields(logrus.Fields{"order side": newOrder.Side,
		"order cp": newOrder.CurrencyPair,
		"挂单卖出数量":   newOrder.Amount,
		"挂单卖出价格":   newOrder.Price,
	}).Info("挂单成功")
	logger.Printf("order created with ID: %s, status: %s\n", createdOrder.Id, createdOrder.Status)
	for createdOrder.Status == "open" {
		order, _, err := client.SpotApi.GetOrder(ctx, createdOrder.Id, createdOrder.CurrencyPair)
		if err != nil {
			panicGateError(err)
		}
		logger.Printf("order %s filled: %s, left: %s\n", order.Id, order.FilledTotal, order.Left)
		logrus.WithFields(logrus.Fields{"order id": order.Id,
			"order FilledTotal": order.FilledTotal,
			"order Left":        order.Left,
			"卖出数量":              order.Amount,
			"卖出价格":              order.Price,
		}).Info("等待挂单完全成交")
		time.Sleep(time.Millisecond * 1500)
		if order.Status == "closed" {
			logger.Printf("order %s filled: %s, left: %s\n", order.Id, order.FilledTotal, order.Left)
			logrus.Info(fmt.Sprintf("order %s filled: %s, left: %s\n", order.Id, order.FilledTotal, order.Left))
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
		logrus.WithFields(logrus.Fields{"order id": t.OrderId,
			"成交数量": t.Amount,
			"成交价格": t.Price,
		}).Info("已成交")
	}
}

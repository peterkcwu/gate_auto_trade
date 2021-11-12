# 使用说明

> 目前支持两个功能, spot购买现货，check查看tickers

`go build`


-func spot/check

-k api-key>

-s api-secret 

-a order-amount 

-p order-price 

-cp currency-pair
 
-a string
        order amount

 -u string
        API based URL used
 
 ## Example
 `gate_auto_trade.exe -func check -k test -s test -cp GT_USDT`
 
`gate_auto_trade.exe -func spot -k test -s test -cp GT_USDT -a 1 -p 6.2`

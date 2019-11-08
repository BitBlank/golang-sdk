package bapp

import (
	"log"
	"testing"
)

var (
	AppKey    = "your_app_key"
	AppSecret = "your_app_secret"
	ReturnUrl = "https://www.google.com"
	NotifyUrl = "https://www.google.com/callback"
)

func TestBAppPay(t *testing.T) {

	s := NewService(AppKey, AppSecret, ReturnUrl, NotifyUrl)
	_, err := s.OrderCreate("123456", "BappTest", 500)
	if err != nil {
		log.Fatalf("OrderCreate failed %+v \n", err)
	}
}

func TestBAppOrderQuery(t *testing.T) {
	s := NewService(AppKey, AppSecret, ReturnUrl, NotifyUrl)
	_, err := s.OrderQuery("b9b71a4ed6559c9fd22443ffba19b5fb")
	if err != nil {
		log.Fatalf("OrderQuery failed %+v \n", err)
	}
}

func TestBAppSign(t *testing.T) {
	request := OrderRequest{
		BAppId:      "20191108072229657c32",
		OrderId:     "073f219de6ade14a3a675b31f5b348da",
		OrderState:  1,
		OrderType:   2,
		Amount:      99,
		AmountType:  "CNY",
		AmountBtc:   1534,
		OrderFee:    1,
		OrderFeeBtc: 16,
		Rate:        6455380,
		CreateTime:  1573197749068,
		PayTime:     1573197753753,
		Body:        "",
		Extra:       "",
		OrderIp:     "",
		Time:        1573197754343,
		AppKey:      "9b3bee8696d9d208",
		Sign:        "8f371ef58906f60b6063d833fb5e333d",
	}
	s := NewService(AppKey, AppSecret, ReturnUrl, NotifyUrl)
	if passed, _ := s.CalcSign(request); !passed {
		log.Fatalf("sign check failed  ")
	}
}

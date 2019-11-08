package bapp

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"
)

const (
	basePath  = "https://bapi.app/api/v2"
	payPath   = basePath + "/pay"
	queryPath = basePath + "/order"
)

type Service struct {
	Api
	appKey    string
	appSecret string
	returnUrl string
	notifyUrl string
}

type Api interface {
	OrderCreate(orderId, body string, amount int) (*PayData, error)
	OrderQuery(orderId string) (*OrderDetail, error)
	CalcSign(request OrderRequest) (bool, error)
}

type PayData struct {
	QrCode string `json:"qr_code"`
	PayUrl string `json:"pay_url"`
}

type PayResponse struct {
	Code int     `json:"code"`
	Msg  string  `json:"msg"`
	Data PayData `json:"data"`
}

type OrderRequest struct {
	BAppId      string `json:"bapp_id"`       //: "20190618171802840b6a",
	OrderId     string `json:"order_id"`      //: "1",
	OrderState  int    `json:"order_state"`   // "order_state": 0: Waiting for user payment 1: Payment success 2: Order timeout is automatically closed
	OrderType   int    `json:"order_type"`    // "order_type": 2,
	Amount      int64  `json:"amount"`        // "amount": 1,
	AmountType  string `json:"amount_type"`   // "amount_type": "CNY",
	AmountBtc   int64  `json:"amount_btc"`    // "amount_btc": 16,
	OrderFee    int64  `json:"order_fee"`     // "order_fee": 0,
	OrderFeeBtc int64  `json:"order_fee_btc"` // "order_fee_btc": 0,
	Rate        int64  `json:"rate"`          // "rate": 6432450,
	CreateTime  int64  `json:"create_time"`   // "create_time": 1560849482796,
	PayTime     int64  `json:"pay_time"`      // "pay_time": 1560859623468,
	Body        string `json:"body"`          // "body": "goods_name",
	Extra       string `json:"extra"`         // "extra": "",
	OrderIp     string `json:"order_ip"`      // "order_ip": "",
	Time        int64  `json:"time"`          // "time": 1561023663119,
	AppKey      string `json:"app_key"`       // "app_key": "4789e57f8629eb9e",
	Sign        string `json:"sign"`          // "sign": "d72e1c8d7efbac64cbc8ec5b76b00671"
}

type OrderDetail struct {
	BAppId     string `json:"bapp_id"`     // "bapp_id": "20190618171802840b6a",
	OrderId    string `json:"order_id"`    // "order_id": "1",
	OrderState int    `json:"order_state"` //"order_state": 1,
	Body       string `json:"body"`        //"body": "goods_name",
	NotifyUrl  string `json:"notify_url"`  //"notify_url": "https://bapi.app/api/experience/notify/test",
	OrderIp    string `json:"order_ip"`    //"order_ip": "",
	Amount     int64  `json:"amount"`      //"amount": 1,
	AmountType string `json:"amount_type"` //"amount_type": "CNY",
	AmountBtc  int64  `json:"amount_btc"`  //"amount_btc": 16,
	PayTime    int64  `json:"pay_time"`    //"pay_time": 1560859623468,
	CreateTime int64  `json:"create_time"` //"create_time": 1560849482796,
	OrderType  int    `json:"order_type"`  //"order_type": 2,
	AppKey     string `json:"app_key"`     //"app_key": "4789e57f8629eb9e",
	Extra      string `json:"extra"`       //"extra": ""
}

type OrderDetailResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data OrderDetail `json:"data"`
}

type PayError string

func NewService(AppKey, AppSecret, rUrl, nUrl string) *Service {
	return &Service{appKey: AppKey, appSecret: AppSecret, returnUrl: rUrl, notifyUrl: nUrl}
}

func (e PayError) Error() string {
	return fmt.Sprintf("Bapp pay error[%s]", string(e))
}

func (s *Service) OrderCreate(orderId, body string, amount int) (*PayData, error) {
	params := map[string]string{
		"order_id":    orderId,
		"amount_type": "CNY",
		"amount":      strconv.Itoa(amount),
		"time":        strconv.FormatInt(time.Now().Unix()*1000, 10),
		"app_key":     s.appKey,
		"body":        body,
		"return_url":  s.returnUrl,
		"notify_url":  s.notifyUrl,
	}
	params["sign"] = signParams(params, s.appSecret)
	var js []byte
	js, err := json.Marshal(params)
	if err != nil {
		return nil, PayError(err.Error())
	}

	resp, err := http.Post(payPath, "application/json; charset=utf-8", bytes.NewReader(js))
	if err != nil {
		return nil, PayError(err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return nil, PayError("StatusCode != 200")
	}
	ret := &PayResponse{}
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, PayError(err.Error())
	}
	if ret.Code != http.StatusOK {
		return nil, PayError(ret.Msg)
	}
	return &ret.Data, nil
}

func (s *Service) OrderQuery(orderId string) (*OrderDetail, error) {
	params := map[string]string{
		"order_id": orderId,
		"time":     strconv.FormatInt(time.Now().Unix(), 10),
		"app_key":  s.appKey,
	}

	var query string
	for k, v := range params {
		query += fmt.Sprintf("%s=%s&", k, v)
	}
	query += fmt.Sprintf("sign=%s", signParams(params, s.appSecret))
	resp, err := http.Get(queryPath + "?" + query)
	if err != nil {
		return nil, PayError(err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return nil, PayError("StatusCode != 200")
	}
	ret := &OrderDetailResponse{}
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, PayError(err.Error())
	}
	if ret.Code != http.StatusOK {
		return nil, PayError(ret.Msg)
	}
	return &ret.Data, nil
}

func (s *Service) CalcSign(request OrderRequest) (bool, error) {
	params, err := convertParams(request)
	if err != nil {
		return false, PayError(err.Error())
	}
	return signParams(params, s.appSecret) == request.Sign, nil
}

// sort params by key-alphabet-accend
func signParams(params map[string]string, appSecret string) string {
	var keys []string
	for k, _ := range params {
		if k != "sign" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var plain string
	for _, k := range keys {
		plain += fmt.Sprintf("%s=%s&", k, params[k])
	}
	plain += "app_secret=" + appSecret
	hash := md5.New()
	hash.Write([]byte(plain))
	return hex.EncodeToString(hash.Sum(nil))
}

func convertParams(request interface{}) (map[string]string, error) {
	var params = make(map[string]string)
	b, err := json.Marshal(request)
	if err != nil {
		return params, err
	}
	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	if err != nil {
		return params, err
	}
	for k, v := range m {
		switch v.(type) {
		case float64:
			params[k] = strconv.FormatInt(int64(v.(float64)), 10)
		case string:
			params[k] = v.(string)
		default:
			fmt.Printf("type %t", v)
		}
	}
	return params, nil
}

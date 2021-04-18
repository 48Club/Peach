package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/CodyGuo/godaemon"
	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {
	b, err := tb.NewBot(tb.Settings{
		Token:  os.Getenv("token"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle(tb.OnQuery, func(q *tb.Query) {
		if strings.ToUpper(q.Text) == "HHHH" {
			results := make(tb.Results, 1)
			results[0] = &tb.PhotoResult{
				URL:       "https://i.imgur.com/JMXhPqI.png",
				ThumbURL:  "https://i.imgur.com/JMXhPqI.png",
				Caption:   "`Pig God: 我发火龙都累死了`",
				ParseMode: tb.ModeMarkdownV2,
			}
			results[0].SetResultID(strconv.Itoa(1))
			_ = b.Answer(q, &tb.QueryResponse{
				Results:   results,
				CacheTime: 1000,
			})
			return
		}
		queryText := strings.Split(strings.ToUpper(q.Text), " ")
		resultsText := "当查询词为空时，默认查询 BTC 汇率"
		if len(queryText) == 0 {
			queryText = append(queryText, "BTC")
		}
		if len(queryText) == 1 {
			if queryText[0] == "" {
				queryText[0] = "BTC"
			} else {
				resultsText = fmt.Sprintf("查询当前 %s 汇率，继续空格输入数量", queryText[0])
			}
			queryText = append(queryText, "1")
		} else if len(queryText) == 2 {
			resultsText = fmt.Sprintf("当前 %s 个 %s 市价", queryText[1], queryText[0])
		} else {
			return
		}
		usdp := getUSDPrice()
		var otherp float64
		for _, v := range []string{"btcusd", " gbpusd", " eurusd", " xrpusd", " ltcusd", " ethusd", " bchusd", " paxusd", " xlmusd", " linkusd", " omgusd", " usdcusd"} {
			if fmt.Sprintf("%susd", strings.ToLower(queryText[0])) == v {
				otherp = getBitstampPrice(v)
				break
			}
		}
		if otherp == 0 {
			otherp = getBinancePrice(queryText[0])
		}
		if otherp == 0 || queryText[0] == "USDT" {
			otherp = getCoingeckoPrice(strings.ToLower(queryText[0]))
		}
		var bc2cp float64
		for _, v := range []string{"usdt", "btc", "busd", "bnb", "eth", "dai"} {
			if strings.ToLower(queryText[0]) == v {
				bc2cp = getBinanceC2CPrice(queryText[0])
				break
			}
		}
		var hc2cp float64

		for k, v := range map[string]string{"btc": "1", "ht": "4", "ltc": "8", "xrp": "7", "eos": "5", "eth": "3", "usdt": "2"} {
			if strings.ToLower(queryText[0]) == k {
				hc2cp = getHuobiC2CPrice(v)
				break
			}
		}
		results := make(tb.Results, 1)
		if usdp > 0 && otherp > 0 {
			count, err := strconv.ParseFloat(queryText[1], 64)
			if err != nil {
				results[0] = &tb.ArticleResult{Title: "货币数量输入错误，", Text: "嘤嘤嘤QAQ"}
				goto errto
			}
			cnyp := ftostring(usdp * otherp * count)
			usdtp := ftostring(otherp * count)
			text := fmt.Sprintf("%s %s = %s USD\n%s %s = %s CNY", queryText[1], queryText[0], usdtp, queryText[1], queryText[0], cnyp)
			text += getC2CStr(count, bc2cp, "币安")
			text += getC2CStr(count, hc2cp, "火币")
			results[0] = &tb.ArticleResult{Title: fmt.Sprintf(resultsText+" %s USD", usdtp), Text: text}
		} else {
			results[0] = &tb.ArticleResult{Title: "暂不支持该货币，", Text: "嘤嘤嘤QAQ"}
		}
	errto:
		results[0].SetResultID(strconv.Itoa(0))

		// results[1] = &tb.AresultsTexticleResult{Title: "赞助我 QAQ!", Text: "USDT(TRC20): `THyvm5rgHWA8D1R89Y12JASVdVhunHNcxd`"}
		// results[0].SetResultID(strconv.Itoa(1))

		err := b.Answer(q, &tb.QueryResponse{
			Results:   results,
			CacheTime: 1,
		})

		if err != nil {
			log.Println(err)
		}
	})
	b.Start()
}

func getC2CStr(count float64, price float64, s string) string {
	if price > 0 {
		c2cpStr := ""
		amount := price * count
		if amount > 100 {
			c2cpStr = fmt.Sprintf("%.2f", amount)
		} else {
			c2cpStr = fmt.Sprintf("%.4f", amount)
		}
		return fmt.Sprintf("\n%s场外价格: %s CNY", s, c2cpStr)
	}
	return ""
}

func ftostring(f float64) string {
	c := 2
	if f < 100 {
		c = 4
	}
	return strconv.FormatFloat(f, 'f', c, 64)
}

type usddata struct {
	Currency string              `json:"currency"`
	Rates    map[string](string) `json:"rates"`
}

type usdres struct {
	Data usddata `json:"data"`
}

func getUSDPrice() float64 {
	var ress usdres
	if err := httpGet("https://api.coinbase.com/v2/exchange-rates?currency=USD", &ress); err == nil {
		if usdPrice, err := strconv.ParseFloat(ress.Data.Rates["CNY"], 64); err == nil {
			return usdPrice
		}
	}
	return 0
}

func getBinancePrice(s string) float64 {
	if s == "USD" {
		return 1
	}
	var price float64
	var ress [][]interface{}
	if err := httpGet(fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%sUSDT&interval=1M&limit=1", s), &ress); !(err != nil || len(ress) < 1 || len(ress[0]) < 5) {
		if price, err = strconv.ParseFloat(ress[0][4].(string), 64); err == nil {
			return price
		}
	}
	return 0
}

func getBinanceC2CPrice(s string) float64 {
	jsonStr := []byte(fmt.Sprintf(`{"page":1,"rows":10,"payTypeList":[],"asset":"%s","tradeType":"SELL","fiat":"CNY"}`, s))
	if resp, err := http.Post("https://c2c.binance.com/gateway-api/v2/public/c2c/adv/search", "application/json", bytes.NewBuffer(jsonStr)); err == nil {
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			resp.Body.Close()
			var res map[string]interface{}
			if err := json.Unmarshal(body, &res); err == nil && res["code"].(string) == "000000" {
				if price, err := strconv.ParseFloat(res["data"].([]interface{})[0].(map[string]interface{})["advDetail"].(map[string]interface{})["price"].(string), 64); err == nil {
					return price
				}
			}
		}
	}
	return 0.0
}

func getHuobiC2CPrice(s string) float64 {
	if resp, err := http.Get(fmt.Sprintf("https://otc-api-hk.eiijo.cn/v1/data/trade-market?coinId=%s&currency=1&tradeType=sell&currPage=1&payMethod=0&acceptOrder=-1&country=&blockType=general&online=1&range=0", s)); err == nil {
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			resp.Body.Close()
			var res map[string]interface{}
			if err := json.Unmarshal(body, &res); err == nil && res["code"].(float64) == 200 {
				if price, err := strconv.ParseFloat(res["data"].([]interface{})[0].(map[string]interface{})["price"].(string), 64); err == nil {
					return price
				}
			}
		}
	}
	return 0.0
}

type SymbolMap struct {
	Id     string `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

var coingecko = []SymbolMap{}
var mapUpdateTime int64

func getCoingeckoPrice(s string) float64 {
	if time.Now().Unix()-mapUpdateTime > 60*60 {
		if err := httpGet("https://api.coingecko.com/api/v3/coins/list", &coingecko); err != nil || len(coingecko) < 1 {
			return 0
		}
		mapUpdateTime = time.Now().Unix()
	}
	var id string
	for _, v := range coingecko {
		if v.Symbol == s {
			id = v.Id
			break
		}
	}
	var price map[string]map[string]float64
	if err := httpGet(fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd", id), &price); err == nil {
		return price[id]["usd"]
	}
	return 0
}

func getBitstampPrice(s string) float64 {
	var price map[string]string
	if err := httpGet(fmt.Sprintf("https://www.bitstamp.net/api/v2/ticker_hour/%s/", s), &price); err == nil {
		if v, ok := price["last"]; ok {
			if p, err := strconv.ParseFloat(v, 64); err == nil {
				return p
			}
		}
	}
	return 0
}

func httpGet(uri string, v interface{}) error {
	var err error
	var resp *http.Response
	if resp, err = http.Get(uri); err == nil {
		var body []byte
		if body, err = ioutil.ReadAll(resp.Body); err == nil {
			resp.Body.Close()
			err = json.Unmarshal(body, &v)
		}
	}
	return err
}

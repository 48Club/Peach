package main

/*
* Contact
* Email: pr@notodom.com
* Sponsor
* ETH/BSC: 0x9F9D41C25c35DA7b87024AF8a04F021DdEfbfFD1
* TRON: TJaCJP66Wh8NsMaCFbNiSDWnwnwufpoLks
* DOGE: D9BXAocUcsrg86bPUM1nGunwfBGFyR36rC
* BTC: bc1qvh5mcucsj45l3v2h7r54dypjz60lxsmfa6daf5
**/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/CodyGuo/godaemon"
	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {

	rawurl := fmt.Sprintf("https://raw.githubusercontent.com/BNB48Club/Peach/%s/file/", GetSha())

	b, err := tb.NewBot(tb.Settings{
		Token:  os.Getenv("token"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}
	answer := func(q *tb.Query, results tb.Results, ct int) {
		results[0].SetResultID(strconv.Itoa(0))
		_ = b.Answer(q, &tb.QueryResponse{
			Results:   results,
			CacheTime: ct,
		})
	}
	inogglist := func(q string, ogglist map[string][]string) (bool, []string) {
		for k, v := range ogglist {
			if strings.EqualFold(q, k) {
				return true, v
			}
		}
		return false, []string{}
	}

	b.Handle(tb.OnQuery, func(q *tb.Query) {
		results := make(tb.Results, 1)
		isinogg, oggv := inogglist(q.Text, ogglist)
		gxq := strings.Split(strings.ToUpper(q.Text), ".GX")
		switch true {
		case strings.ToUpper(q.Text) == "HHHH":
			results[0] = &tb.PhotoResult{
				URL:       fmt.Sprintf("%sJMXhPqI.png", rawurl),
				ThumbURL:  fmt.Sprintf("%sJMXhPqI.png", rawurl),
				Caption:   "`Pig God: 我发火龙都累死了`",
				ParseMode: tb.ModeMarkdownV2,
			}
			answer(q, results, 60)
			return
		case isinogg:
			results[0] = &tb.VoiceResult{
				URL:   fmt.Sprintf("%s%s", rawurl, oggv[0]),
				Title: oggv[1],
			}
			answer(q, results, 60)
			return
		case len(gxq) == 2 && gxq[0] == "":
			gxindex := int64(0)
			if gxq[1] == "" {
				rand.Seed(time.Now().UnixNano())
				gxindex = rand.Int63n(int64(len(gxlist)))
			} else {
				var err error
				gxindex, err = strconv.ParseInt(gxq[1], 10, 64)
				if err != nil || gxindex > int64(len(gxlist)-1) {
					results[0] = &tb.ArticleResult{
						Title: "暂不支持该货币，", Text: "嘤嘤嘤QAQ",
					}
					answer(q, results, 1)
					return
				}
			}
			results[0] = &tb.VoiceResult{
				URL:   fmt.Sprintf("%sgx/%d.ogg", rawurl, gxindex),
				Title: gxlist[gxindex],
			}
			answer(q, results, 1)
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
		answer(q, results, 1)

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

func GetSha() string {
	res := []map[string]interface{}{}
	err := httpGet("https://api.github.com/repos/BNB48Club/Peach/commits", &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	return res[0]["sha"].(string)
}

var (
	ogglist = map[string][]string{
		"0000":   {"0000.ogg", "归零"},
		".kong":  {"kong.ogg", "直接重仓空进去"},
		".suoha": {"suoha.ogg", "已经在谷底了，梭！"},
		".jg":    {"jigou.ogg", "机构进场了，抄底！"},
		".jc":    {"加仓之歌.ogg", "买的多，赢得多，可以单车变摩托！"},
	}
	gxlist = []string{
		/* 0 */ "八点尊",
		/* 1 */ "把那个消息撤回去",
		/* 2 */ "不是我针对谁，在座的各位都是我儿子",
		/* 3 */ "不要放DJ了，几十岁的人了",
		/* 4 */ "不要聊了先上DJ",
		/* 5 */ "不要挑战权威",
		/* 6 */ "道不同不相为谋",
		/* 7 */ "等下你会被莫名其妙移出该群的",
		/* 8 */ "等着猝死把，我先睡了么么哒",
		/* 9 */ "搞的自己很忙一样",
		/* 10 */ "给钱给钱，红包过来什么都有",
		/* 11 */ "滚一边去",
		/* 12 */ "几百条消息没有一条是关于我的",
		/* 13 */ "加个微信有这么难嘛",
		/* 14 */ "来到这个群不要泡群里面的妹子",
		/* 15 */ "来点DJ啊",
		/* 16 */ "老子听到我的语音就烦",
		/* 17 */ "没有，滚",
		/* 18 */ "每次喔都会找话题插一下",
		/* 19 */ "你何德何能让我加你好友啊",
		/* 20 */ "你们聊啊，我插不了嘴的",
		/* 21 */ "你们这群表面群友",
		/* 22 */ "你能不能不要在这里恶心啊",
		/* 23 */ "泡一杯红茶来喝一下",
		/* 24 */ "去跟张学友称兄掉地啊",
		/* 25 */ "群里的妹子有没有甜言蜜语的",
		/* 26 */ "群里面只有妹子能艾特我",
		/* 27 */ "群主把楼上这个叼毛踢掉",
		/* 28 */ "人家说要个鸡脖你给不给啊？",
		/* 29 */ "睡觉啦，不要在群里面发我的语音",
		/* 30 */ "天籁",
		/* 31 */ "晚上发点片片看啊",
		/* 32 */ "文明，wenming",
		/* 33 */ "我今晚那个炒米粉赚回来了",
		/* 34 */ "我们群主好搞笑啊",
		/* 35 */ "我是做鸭的",
		/* 36 */ "我说你们的微信小助手啊贤",
		/* 37 */ "下面我给大家带来一首英文歌",
		/* 38 */ "先来首DJ有那么难嘛",
		/* 39 */ "小姐姐我还是单身喔",
		/* 40 */ "笑死我了tmd，哎呦",
		/* 41 */ "新进群的妹子加我一下",
		/* 42 */ "一个文明的聊天群搞的乌鸦胀气的",
		/* 43 */ "有没有还有煞笔没有睡觉啊，出来聊天啊",
		/* 44 */ "有没有人知道那个当当当是什么音乐",
		/* 45 */ "在bb老子踢你",
		/* 46 */ "早上好兄弟姐妹们",
		/* 47 */ "这个群是怎么了，是感情纠纷",
		/* 48 */ "这是人说的话吗",
		/* 49 */ "真不知道你们一天到晚聊什么",
		/* 50 */ "左右为难啊",
		/* 51 */ "DJ在哪里不知道打电话找他",
		/* 52 */ "duang~",
	}
)

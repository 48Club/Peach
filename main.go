package main

/*
* Contact
* Email: pr@notodom.com
* Sponsor
* ETH/BSC: 0x8888865ca6D38365d49e63098ceB37D48Fe88888
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
	"github.com/google/uuid"
	tb "gopkg.in/telebot.v3"
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
	answer := func(q *tb.Query, results tb.Results, ct int) error {
		results[0].SetResultID(strconv.Itoa(0))
		return b.Answer(q, &tb.QueryResponse{
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

	b.Handle(tb.OnQuery, func(cont tb.Context) error {
		q := cont.Query()
		results := make(tb.Results, 1)
		isinogg, oggv := inogglist(q.Text, ogglist)
		gxq := strings.Split(strings.ToUpper(q.Text), ".GX")
		switch true {
		case strings.ToUpper(q.Text) == "HHHH":
			results[0] = &tb.PhotoResult{
				URL:        fmt.Sprintf("%sJMXhPqI.png", rawurl),
				ThumbURL:   fmt.Sprintf("%sJMXhPqI.png", rawurl),
				Caption:    "`Pig God: 我发火龙都累死了`",
				ResultBase: tb.ResultBase{ParseMode: tb.ModeMarkdownV2},
			}
			return answer(q, results, 60)

		case isinogg:
			results[0] = &tb.VoiceResult{
				URL:   fmt.Sprintf("%s%s", rawurl, oggv[0]),
				Title: oggv[1],
			}
			return answer(q, results, 60)

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
					return answer(q, results, 1)
				}
			}
			results[0] = &tb.VoiceResult{
				URL:   fmt.Sprintf("%sgx/%d.ogg", rawurl, gxindex),
				Title: gxlist[gxindex],
			}
			return answer(q, results, 1)

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
			return nil
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

		if usdp > 0 && otherp > 0 {
			count, err := strconv.ParseFloat(queryText[1], 64)
			if err != nil {
				results[0] = &tb.ArticleResult{Title: "货币数量输入错误，", Text: "嘤嘤嘤QAQ"}
				goto errto
			}
			var (
				bc2cp    float64
				userinfo string
				sellT    string
				jumpU    bool
			)
			for _, v := range []string{"usdt", "btc", "eth"} {
				if strings.ToLower(queryText[0]) == v {
					bc2cp, userinfo, sellT = getBinanceC2CPrice(queryText[0], usdp*otherp*count)
					break
				}
			}
			if jumpU = bc2cp == 0; jumpU {
				bc2cp, userinfo, sellT = getBinanceC2CPrice("USDT", usdp*otherp*count)
			}
			cnyp := ftostring(usdp * otherp * count)
			usdtp := ftostring(otherp * count)
			text := fmt.Sprintf("%s %s ≈ %s USD ≈ %s CNY", queryText[1], queryText[0], usdtp, cnyp)
			var lines [][]tb.InlineButton
			switch_inline_query := []tb.InlineButton{{Text: "我也试试", InlineQuery: fmt.Sprintf("%s %s", queryText[0], queryText[1])}}
			lines = [][]tb.InlineButton{}
			if bc2cp > 0 {
				if jumpU {
					text += fmt.Sprintf("\n\nPexPay: %s => USDT ≈> %.2f CNY", queryText[0], bc2cp*count*otherp)
				} else {
					text += fmt.Sprintf("\n\nPexPay: %s ≈> %.2f CNY", queryText[0], bc2cp*count)
				}
				thisUuid := uuid.NewString()
				calllist[thisUuid] = []string{text, userinfo, "SELL", fmt.Sprintf("%d", q.Sender.ID), fmt.Sprintf("%s %s", queryText[0], queryText[1]), sellT}
				lines = append(lines, []tb.InlineButton{{Text: "查询场外", Data: thisUuid}})
			}
			// text += "\n\n🪧 底部常驻广告位招租 @elrepo"
			lines = append(lines, switch_inline_query)
			results[0] = &tb.ArticleResult{
				Title: fmt.Sprintf(resultsText+" %s USD", usdtp),
				Text:  text,
				ResultBase: tb.ResultBase{
					ReplyMarkup: &tb.ReplyMarkup{
						InlineKeyboard: lines,
					},
				},
			}
		} else {
			results[0] = &tb.ArticleResult{Title: "暂不支持该货币，", Text: "嘤嘤嘤QAQ"}
		}
	errto:
		results[0].SetResultID(strconv.Itoa(0))
		return answer(q, results, 1)

	})
	b.Handle(tb.OnCallback, func(cont tb.Context) error {
		call := cont.Callback()
		switch_inline_query := []tb.InlineButton{{Text: "我也试试", InlineQuery: ""}}
		data, ok := calllist[call.Data]
		if !ok {
			_ = b.Respond(call, &tb.CallbackResponse{
				Text:      "报价失效咯~ 请重新发起查询",
				ShowAlert: true,
			})
			_, err := b.EditReplyMarkup(call.Message, &tb.ReplyMarkup{
				InlineKeyboard: [][]tb.InlineButton{switch_inline_query},
			})
			return err
		}
		if len(data) == 0 {
			return nil
		}
		switch data[2] {
		case "SELL":
			if fmt.Sprintf("%d", call.Sender.ID) != data[3] { // 此报价不是你发起的哦~
				return b.Respond(call, &tb.CallbackResponse{
					Text:      "此报价不是你发起的哦~",
					ShowAlert: true,
				})
			}
			calllist[call.Data] = []string{}
			userNo := data[1]
			rt := ""
			if userInfo := getUserInfo(userNo); userInfo.Code == "000000" {
				rt += fmt.Sprintf("\n商户: %s (保证金 %.2f %s)", userInfo.Data.UserDetailVo.NickName, userInfo.Data.UserDetailVo.DepositAmount, userInfo.Data.UserDetailVo.DepositCurrency)
				KYC := []string{}
				cKyc := func(v bool, c string) {
					if v {
						KYC = append(KYC, c)
					}
				}
				cKyc(userInfo.Data.UserDetailVo.EmailVerified, "邮箱")
				cKyc(userInfo.Data.UserDetailVo.BindMobile, "手机")
				cKyc(userInfo.Data.UserDetailVo.KycVerified, "身份认证")
				rt += fmt.Sprintf("\nKYC: %s", strings.Join(KYC[:], "+"))
				rt += fmt.Sprintf("\n方式: %s", data[5])
				rt += fmt.Sprintf("\n成交: 总 %.f 单, 月 %.f 单, 成交率%.2f%%", userInfo.Data.UserDetailVo.UserStatsRet.CompletedOrderNum, userInfo.Data.UserDetailVo.UserStatsRet.CompletedOrderNumOfLatest30day, userInfo.Data.UserDetailVo.UserStatsRet.FinishRateLatest30day*100)
				rt += fmt.Sprintf("\n付款: 平均 %.2f 分放行, %.2f 分付款", userInfo.Data.UserDetailVo.UserStatsRet.AvgReleaseTimeOfLatest30day/60, userInfo.Data.UserDetailVo.UserStatsRet.AvgPayTimeOfLatest30day/60)
				rt += fmt.Sprintf("\n账户: 已注册 %.f 天; 首次交易于 %.f 天前", userInfo.Data.UserDetailVo.UserStatsRet.RegisterDays, userInfo.Data.UserDetailVo.UserStatsRet.FirstOrderDays)
				line := []tb.InlineButton{{Text: "前往交易", URL: fmt.Sprintf("https://www.pexpay.com/zh-CN/advertiserDetail?advertiserNo=%s", data[1])}}
				switch_inline_query = []tb.InlineButton{{Text: "我也试试", InlineQuery: data[4]}}
				if _, err := b.Edit(call.Message, data[0]+rt+"\n\n🪧 底部常驻广告位招租 @elrepo", &tb.ReplyMarkup{
					InlineKeyboard: [][]tb.InlineButton{line, switch_inline_query},
				}); err == nil {
					delete(calllist, call.Data)
					return err
				}
			}
		}
		calllist[call.Data] = data
		return nil
	})
	b.Start()
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

type TradeInfo struct {
	Code  string
	Data  []TradeData
	Total float64
}
type TradeData struct {
	AdDetailResp AdDetailResp
	AdvertiserVo AdvertiserVo
}
type AdDetailResp struct {
	Price        string
	TradeMethods []TradeMethods
}
type AdvertiserVo struct {
	UserNo string
}
type TradeMethods struct {
	TradeMethodShortName string
}

func getBinanceC2CPrice(s string, amount float64) (float64, string, string) {
	jsonStr := []byte(fmt.Sprintf(`{"page":1,"rows":10,"payTypes":[],"classifies":[],"asset":"%s","tradeType":"SELL","fiat":"CNY","publisherType":null,"filter":{"payTypes":[]},"transAmount":"%.2f"}`, s, amount))

	reqest, _ := http.NewRequest("POST", "https://www.pexpay.com/bapi/c2c/v1/friendly/c2c/ad/search", bytes.NewBuffer(jsonStr))
	reqest.Header.Add("content-type", "application/json")
	reqest.Header.Add("lang", "zh-CN")

	if resp, err := http.DefaultClient.Do(reqest); err == nil {
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			defer resp.Body.Close()
			var tradeInfo TradeInfo
			if err := json.Unmarshal(body, &tradeInfo); err == nil && tradeInfo.Code == "000000" && tradeInfo.Total > 0 {
				first := tradeInfo.Data[0]
				if price, err := strconv.ParseFloat(first.AdDetailResp.Price, 64); err == nil {
					userNo := first.AdvertiserVo.UserNo
					bType := []string{}
					for _, v := range first.AdDetailResp.TradeMethods {
						bType = append(bType, v.TradeMethodShortName)

					}
					return price, userNo, strings.Join(bType[:], "+")
				}
			}
		}
	}
	return 0.0, "", ""
}

type UserInfo struct {
	Code string
	Data UserData
}
type UserData struct {
	UserDetailVo UserDetailVo
}
type UserDetailVo struct {
	NickName        string
	DepositAmount   float64
	DepositCurrency string
	EmailVerified   bool
	BindMobile      bool
	KycVerified     bool
	UserStatsRet    UserStatsRet
}
type UserStatsRet struct {
	RegisterDays                   float64
	FirstOrderDays                 float64
	AvgReleaseTimeOfLatest30day    float64
	AvgPayTimeOfLatest30day        float64
	FinishRateLatest30day          float64
	CompletedOrderNum              float64
	CompletedOrderNumOfLatest30day float64
}

func getUserInfo(userNo string) UserInfo {
	var userInfo UserInfo
	if resp, err0 := http.Get(fmt.Sprintf("https://www.pexpay.com/bapi/c2c/v1/friendly/c2c/user/profile-and-ads-list?userNo=%s", userNo)); err0 == nil {
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			defer resp.Body.Close()
			_ = json.Unmarshal(body, &userInfo)
		}
	}
	return userInfo
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
	calllist = map[string][]string{}
	ogglist  = map[string][]string{
		"0000":     {"0000.ogg", "归零"},
		".kong":    {"kong.ogg", "直接重仓空进去"},
		".suoha":   {"suoha.ogg", "已经在谷底了，梭！"},
		".jg":      {"jigou.ogg", "机构进场了，抄底！"},
		".jc":      {"加仓之歌.ogg", "买的多，赢得多，可以单车变摩托！"},
		".xinyang": {"信仰.ogg", "你没有信仰的话，你就会错过暴富的机会，懂不懂啊？"},
		".jihui":   {"机会.ogg", "你还要错过多少次机会？人生会给你多少次机会？拿出胆子来，抄底！操!"},
		".huit":    {"回调.ogg", "现在回调，是给你最后的机会加仓了，知不知道啊？"},
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

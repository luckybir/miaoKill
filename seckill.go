package main

import (
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func secondKill() {
	checkLogin()
	requestSecKillURL()
	requestSecKillCheckoutPage()
	submitSecKillOrder()
}

func requestSecKillURL() {
	//"""访问商品的抢购链接（用于设置cookie等"""

	getSkuTittle()
	getUserName()

	waitSecKillStart()
	getSecKillURL()
	navigateSecKillURL()
}

func waitSecKillStart() {
	now := time.Now()
	t := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, time.Local)

	Sugar.Info("waiting...")

	for {
		// 本地时间减去与京东的时间差，能够将时间误差提升到0.1秒附近
		// 具体精度依赖获取京东服务器时间的网络时间损耗
		if time.Now().Unix()*1000+secKillInfo.basic.serverTimeOffset > t.Unix()*1000 {
			Sugar.Info("second kill time arrive……")
			break
		} else {
			time.Sleep(500)
		}
	}

}

func getUserName() {
	req, err := http.NewRequest("GET", "https://passport.jd.com/user/petName/getUserInfoForMiniJd.action", nil)
	if err != nil {
		Sugar.Error(err)
	}

	req.Header.Set("User-Agent", secKillInfo.basic.userAgent)
	req.Header.Set("Referer", "https://order.jd.com/center/list.action")

	query := req.URL.Query()

	callback := "jQuery" + strconv.Itoa(randomInt(1000000, 9999999))
	query.Add("callback", callback)
	query.Add("_", string(time.Now().Unix()*1000))
	req.URL.RawQuery = query.Encode()

	resp, err := secKillInfo.basic.client.Do(req)
	if err != nil {
		Sugar.Error(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Sugar.Error(err)
	}

	r := regexp.MustCompile(`"nickName":"([^"]*)",`)
	if r.Match(body) {
		Sugar.Infof("welcome %v", r.FindStringSubmatch(string(body))[1])
	}

}

func getSecKillURL() {
	//"""获取商品的抢购链接
	//点击"抢购"按钮后，会有两次302跳转，最后到达订单结算页面
	//这里返回第一次跳转后的页面url，作为商品的抢购链接
	//:return: 商品的抢购链接
	//"""
	//url = 'https://itemko.jd.com/itemShowBtn'
	//payload = {
	//	'callback': 'jQuery{}'.format(random.randint(1000000, 9999999)),
	//		'skuId': self.sku_id,
	//		'from': 'pc',
	//		'_': str(int(time.time() * 1000)),
	//}
	//headers = {
	//	'User-Agent': self.user_agent,
	//		'Host': 'itemko.jd.com',
	//		'Referer': 'https://item.jd.com/{}.html'.format(self.sku_id),
	//}
	//while True:
	//resp = self.session.get(url=url, headers=headers, params=payload)
	//resp_json = parse_json(resp.text)
	//if resp_json.get('url'):
	//# https://divide.jd.com/user_routing?skuId=8654289&sn=c3f4ececd8461f0e4d7267e96a91e0e0&from=pc
	//router_url = 'https:' + resp_json.get('url')
	//# https://marathon.jd.com/captcha.html?skuId=8654289&sn=c3f4ececd8461f0e4d7267e96a91e0e0&from=pc
	//seckill_url = router_url.replace(
	//	'divide', 'marathon').replace(
	//	'user_routing', 'captcha.html')
	//logger.info("抢购链接获取成功: %s", seckill_url)
	//return seckill_url
	//else:
	//logger.info("抢购链接获取失败，稍后自动重试")
	//wait_some_time()
	Sugar.Info("get second kill URL")

	var routerURL string

	for i := 0; i < 60; i++ {

		req, err := http.NewRequest("GET", "https://itemko.jd.com/itemShowBtn", nil)
		if err != nil {
			Sugar.Error(err)
		}

		req.Header.Set("User-Agent", secKillInfo.basic.userAgent)
		req.Header.Set("Referer", "https://item.jd.com/100012043978.html")

		query := req.URL.Query()

		callback := "jQuery" + strconv.Itoa(randomInt(1000000, 9999999))
		query.Add("callback", callback)
		query.Add("skuId", secKillInfo.basic.skuID)
		query.Add("from", "pc")
		query.Add("_", string(time.Now().Unix()*1000))
		req.URL.RawQuery = query.Encode()

		resp, err := secKillInfo.basic.client.Do(req)
		if err != nil {
			Sugar.Error(err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Sugar.Error(err)
		}

		resp.Body.Close()

		r := regexp.MustCompile(`"url":"([^"]+)"`)
		if r.Match(body) {
			routerURL = "https:" + r.FindStringSubmatch(string(body))[1]
			break
		} else {
			Sugar.Infof("fail: %s",body)
			waitRandomTime()
		}

	}

	// "url":"//divide.jd.com/user_routing?skuId=100012043978&sn=7a5b750fa95e115d8536d816e407b50f&from=pc"})
	if routerURL != "" {
		r := regexp.MustCompile(`user_routing\?(.+)`)

		if r.MatchString(routerURL) {
			secKillInfo.secKill.URL = "https://marathon.jd.com/captcha.html?" + r.FindStringSubmatch(routerURL)[1]
		}

		Sugar.Info("get second kill URL successful")
	} else {
		Sugar.Fatal("get second kill URL failed")
	}

}

func navigateSecKillURL() {
	Sugar.Info("navigate second kill URL")

	req, err := http.NewRequest("GET", "https://itemko.jd.com/itemShowBtn", nil)
	if err != nil {
		Sugar.Error(err)
	}

	req.Header.Set("User-Agent", secKillInfo.basic.userAgent)
	req.Header.Set("Referer", "https://item.jd.com/100012043978.html")
	req.Header.Set("Host", "marathon.jd.com")

	resp, err := secKillInfo.basic.client.Do(req)
	if err != nil {
		Sugar.Error(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Sugar.Error(err)
	}

	Sugar.Info(string(body))

}

func requestSecKillCheckoutPage() {

	Sugar.Info("访问抢购订单结算页面...")

	req, err := http.NewRequest("GET", "https://marathon.jd.com/seckill/seckill.action", nil)
	if err != nil {
		Sugar.Fatal(err)
	}

	query := req.URL.Query()
	query.Add("skuId", secKillInfo.basic.skuID)
	query.Add("num", "2") //		'num': self.seckill_num
	query.Add("rid", string(time.Now().Unix()))
	req.URL.RawQuery = query.Encode()

	req.Header.Set("User-Agent", secKillInfo.basic.userAgent)
	req.Header.Set("Referer", "https://item.jd.com/100012043978.html")
	req.Header.Set("Host", "marathon.jd.com")

	resp, err := secKillInfo.basic.client.Do(req)
	if err != nil {
		Sugar.Error(err)
	}
	defer resp.Body.Close()

	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	Sugar.Error(err)
	//}

	//Sugar.Info(string(body))
}

func submitSecKillOrder() {
	//	"""提交抢购（秒杀）订单
	//	:return: 抢购结果 True/False

	reqBody := getSecKillOrderData()

	for {
		Sugar.Info("提交抢购订单...")
		req, err := http.NewRequest(http.MethodPost, "https://marathon.jd.com/seckillnew/orderService/pc/submitOrder.action", strings.NewReader(reqBody))
		if err != nil {
			Sugar.Fatal(err)
		}

		query := req.URL.Query()
		query.Add("skuId", secKillInfo.basic.skuID)
		req.URL.RawQuery = query.Encode()

		req.Header.Set("User-Agent", secKillInfo.basic.userAgent)

		referer := "https://marathon.jd.com/seckill/seckill.action?skuId=" + secKillInfo.basic.skuID + "&num=2&rid=" + strconv.FormatInt(time.Now().Unix(), 10)
		req.Header.Set("Referer", referer)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := secKillInfo.basic.client.Do(req)
		if err != nil {
			Sugar.Error(err)
		}

		//	# 返回信息
		//	# 抢购失败：
		//	# {'errorMessage': '很遗憾没有抢到，再接再厉哦。', 'orderId': 0, 'resultCode': 60074, 'skuId': 0, 'success': False}
		//	# {'errorMessage': '抱歉，您提交过快，请稍后再提交订单！', 'orderId': 0, 'resultCode': 60017, 'skuId': 0, 'success': False}
		//	# {'errorMessage': '系统正在开小差，请重试~~', 'orderId': 0, 'resultCode': 90013, 'skuId': 0, 'success': False}
		//	# 抢购成功：
		//	# {"appUrl":"xxxxx","orderId":820227xxxxx,"pcUrl":"xxxxx","resultCode":0,"skuId":0,"success":true,"totalMoney":"xxxxx"}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Sugar.Error(err)
		}

		success := gjson.GetBytes(body, "success").Bool()
		if success {
			orderID := gjson.GetBytes(body, "orderId").String()
			totalMoney := gjson.GetBytes(body, "totalMoney").String()
			payURL := "https:" + gjson.GetBytes(body, "pcUrl").String()
			Sugar.Infof("抢购成功，订单号:%v, 总价:%v, 电脑端付款链接:%v", orderID, totalMoney, payURL)
			Sugar.Infof("resp: %s", body)
			break
		} else {
			Sugar.Infof("抢购失败，返回信息:%s", body)
			waitRandomTime()
		}

		resp.Body.Close()
	}

}

func getSecKillOrderData() string {
	Sugar.Info("生成提交抢购订单所需参数...")
	// 获取用户秒杀初始化信息

	//	"""获取秒杀初始化信息（包括：地址，发票，token）
	//	:return: 初始化信息组成的dict
	//	"""
	Sugar.Info("获取秒杀初始化信息...")

	data := url.Values{}
	data.Set("sku", secKillInfo.basic.skuID)
	data.Set("num", "2")
	data.Set("isModifyAddress", "false")

	req, err := http.NewRequest(http.MethodPost, "https://marathon.jd.com/seckillnew/orderService/pc/init.action", strings.NewReader(data.Encode()))

	if err != nil {
		Sugar.Fatal(err)
	}

	req.Header.Set("User-Agent", secKillInfo.basic.userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := secKillInfo.basic.client.Do(req)
	if err != nil {
		Sugar.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	data = url.Values{}
	data.Set("skuId", secKillInfo.basic.skuID)
	data.Set("num", "2")

	v := gjson.GetBytes(body, "addressList.0.id").String()
	data.Set("addressId", v)

	data.Set("yuShou", "true")
	data.Set("isModifyAddress", "false")

	v = gjson.GetBytes(body, "addressList.0.name").String()
	data.Set("name", v)

	v = gjson.GetBytes(body, "addressList.0.provinceId").String()
	data.Set("provinceId", "false")

	v = gjson.GetBytes(body, "addressList.0.cityId").String()
	data.Set("cityId", v)

	v = gjson.GetBytes(body, "addressList.0.countyId").String()
	data.Set("countyId", v)

	v = gjson.GetBytes(body, "addressList.0.townId").String()
	data.Set("townId", v)

	v = gjson.GetBytes(body, "addressList.0.addressDetail").String()
	data.Set("addressDetail", v)

	v = gjson.GetBytes(body, "addressList.0.mobile").String()
	data.Set("mobile", "false")

	v = gjson.GetBytes(body, "addressList.0.mobileKey").String()
	data.Set("mobileKey", v)

	v = gjson.GetBytes(body, "addressList.0.email").String()
	data.Set("email", v)

	data.Set("postCode", "")

	v = gjson.GetBytes(body, "invoiceInfo.invoiceTitle").String()
	data.Set("invoiceTitle", v)

	v = gjson.GetBytes(body, "invoiceInfo.invoiceCompanyName").String()
	data.Set("invoiceCompanyName", v)

	v = gjson.GetBytes(body, "invoiceInfo.invoiceCompanyName").String()
	data.Set("invoiceContent", v)

	data.Set("invoiceTaxpayerNO", "")
	data.Set("invoiceEmail", "")

	v = gjson.GetBytes(body, "invoiceInfo.invoicePhone").String()
	data.Set("invoicePhone", v)

	v = gjson.GetBytes(body, "invoiceInfo.invoicePhoneKey").String()
	data.Set("invoicePhoneKey", v)

	data.Set("invoice", "false") // if invoice_info else 'false',
	data.Set("password", "")     //  global_config.get('account', 'payment_pwd'),
	data.Set("codTimeType", "3")
	data.Set("paymentType", "4")
	data.Set("areaCode", "")
	data.Set("overseas", "0")
	data.Set("phone", "")
	data.Set("eid", "") //global_config.getRaw('config', 'eid'),
	data.Set("fp", "")  //global_config.getRaw('config', 'fp'),

	v = gjson.GetBytes(body, "token").String()
	data.Set("token", v)

	data.Set("pru", "")

	return data.Encode()

}

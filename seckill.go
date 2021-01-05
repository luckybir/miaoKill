package main

import (
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"math/rand"
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
		if time.Now().Unix()*1000+loginInfo.serverTimeOffset > t.Unix()*1000 {
			Sugar.Info("时间到达，开始执行……")
			break
		} else {
			time.Sleep(500)
		}
	}

}

func getUserName() {
	url := `https://passport.jd.com/user/petName/getUserInfoForMiniJd.action`

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		Sugar.Error(err)
	}

	req.Header.Set("User-Agent", loginInfo.userAgent)
	req.Header.Set("Referer", "https://order.jd.com/center/list.action")

	query := req.URL.Query()

	rand.Seed(time.Now().Unix())
	callback := "jQuery" + strconv.Itoa(rand.Intn(8999998)+1000000)
	query.Add("callback", callback)
	query.Add("_", string(time.Now().Unix()*1000))
	req.URL.RawQuery = query.Encode()

	resp, err := loginInfo.client.Do(req)
	if err != nil {
		Sugar.Error(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Sugar.Error(err)
	}

	r := regexp.MustCompile(`"nickName":"(.*)",`)
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
	Sugar.Info("获取抢购URL")

	var routerURL string

	for i := 0; i < 10; i++ {
		url := `https://itemko.jd.com/itemShowBtn`

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			Sugar.Error(err)
		}

		req.Header.Set("User-Agent", loginInfo.userAgent)
		req.Header.Set("Referer", "https://item.jd.com/100012043978.html")

		query := req.URL.Query()

		rand.Seed(time.Now().Unix())
		callback := "jQuery" + strconv.Itoa(rand.Intn(8999998)+1000000)
		query.Add("callback", callback)
		query.Add("skuId", loginInfo.skuID)
		query.Add("from", "pc")
		query.Add("_", string(time.Now().Unix()*1000))
		req.URL.RawQuery = query.Encode()

		resp, err := loginInfo.client.Do(req)
		if err != nil {
			Sugar.Error(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Sugar.Error(err)
		}

		r := regexp.MustCompile(`"url":"([^"]+)"`)
		if r.Match(body) {
			routerURL = "https:" + r.FindStringSubmatch(string(body))[1]
			//Sugar.Infof("seckill URL:%v", loginInfo.secKillURL)
			break
		} else {
			Sugar.Info(string(body))
			//Sugar.Fatal(string(body))
			//Sugar.Fatal("get second kill URL failure")
			waitRandomTime()
		}

	}

	// "url":"//divide.jd.com/user_routing?skuId=100012043978&sn=7a5b750fa95e115d8536d816e407b50f&from=pc"})
	if routerURL != "" {
		r := regexp.MustCompile(`user_routing\?(.+)`)

		if r.MatchString(routerURL) {
			loginInfo.secKillURL = "https://marathon.jd.com/captcha.html?" + r.FindStringSubmatch(routerURL)[1]
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

	req.Header.Set("User-Agent", loginInfo.userAgent)
	req.Header.Set("Referer", "https://item.jd.com/100012043978.html")
	req.Header.Set("Host", "marathon.jd.com")

	resp, err := loginInfo.client.Do(req)
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
	query.Add("skuId", loginInfo.skuID)
	query.Add("num", "2") //		'num': self.seckill_num
	query.Add("rid", string(time.Now().Unix()))
	req.URL.RawQuery = query.Encode()

	req.Header.Set("User-Agent", loginInfo.userAgent)
	req.Header.Set("Referer", "https://item.jd.com/100012043978.html")
	req.Header.Set("Host", "marathon.jd.com")

	resp, err := loginInfo.client.Do(req)
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
		query.Add("skuId", loginInfo.skuID)
		req.URL.RawQuery = query.Encode()

		req.Header.Set("User-Agent", loginInfo.userAgent)

		referer := "https://marathon.jd.com/seckill/seckill.action?skuId=" + loginInfo.skuID + "&num=2&rid=" + strconv.FormatInt(time.Now().Unix(), 10)
		req.Header.Set("Referer", referer)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := loginInfo.client.Do(req)
		if err != nil {
			Sugar.Error(err)
		}
		defer resp.Body.Close()

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

		}
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
	data.Set("sku", loginInfo.skuID)
	data.Set("num", "2")
	data.Set("isModifyAddress", "false")

	req, err := http.NewRequest(http.MethodPost, "https://marathon.jd.com/seckillnew/orderService/pc/init.action", strings.NewReader(data.Encode()))

	if err != nil {
		Sugar.Fatal(err)
	}

	req.Header.Set("User-Agent", loginInfo.userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := loginInfo.client.Do(req)
	if err != nil {
		Sugar.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))

	Sugar.Fatal(string(body))

	//initInfo := gjson.Get(string(body),loginInfo.skuID)

	//self.seckill_init_info[self.sku_id] = self._get_seckill_init_info()
	//init_info = self.seckill_init_info.get(self.sku_id)

	//default_address = init_info['addressList'][0]  # 默认地址dict
	//invoice_info = init_info.get('invoiceInfo', {})  # 默认发票信息dict, 有可能不返回
	//token = init_info['token']

	data = url.Values{}
	data.Set("skuId", loginInfo.skuID)
	data.Set("num", "2")
	data.Set("addressId", "2") //		'addressId': default_address['id'],
	data.Set("yuShou", "true")
	data.Set("isModifyAddress", "false")
	data.Set("name", "false")          //		'name': default_address['name'],
	data.Set("provinceId", "false")    //		'provinceId': default_address['provinceId'],
	data.Set("cityId", "false")        // default_address['cityId'],
	data.Set("countyId", "false")      //  default_address['countyId'],
	data.Set("townId", "false")        // default_address['townId'],
	data.Set("addressDetail", "false") // default_address['addressDetail'],
	data.Set("mobile", "false")        // default_address['mobile'],
	data.Set("mobileKey", "false")     // default_address['mobileKey'],
	data.Set("email", "false")         // default_address.get('email', ''),
	data.Set("postCode", "")
	data.Set("invoiceTitle", "") //invoice_info.get('invoiceTitle', -1),
	data.Set("invoiceCompanyName", "")
	data.Set("invoiceContent", "") //invoice_info.get('invoiceContentType', 1),
	data.Set("invoiceTaxpayerNO", "")
	data.Set("invoiceEmail", "")
	data.Set("invoicePhone", "")    //invoice_info.get('invoicePhone', ''),
	data.Set("invoicePhoneKey", "") //invoice_info.get('invoicePhoneKey', ''),
	data.Set("invoice", "true")     // if invoice_info else 'false',
	data.Set("password", "")        //  global_config.get('account', 'payment_pwd'),
	data.Set("codTimeType", "3")
	data.Set("paymentType", "4")
	data.Set("areaCode", "")
	data.Set("overseas", "0")
	data.Set("phone", "")
	data.Set("eid", "")        //global_config.getRaw('config', 'eid'),
	data.Set("fp", "")         //global_config.getRaw('config', 'fp'),
	data.Set("token", "token") //token
	data.Set("pru", "")        // init_info['token']

	return data.Encode()

}

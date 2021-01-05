package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

func secondKill() {
	checkLogin()
	requestSecKillURL()

}

func requestSecKillURL() {
	//"""访问商品的抢购链接（用于设置cookie等"""
	//logger.info('用户:{}'.format(self.get_username()))
	//logger.info('商品名称:{}'.format(self.get_sku_title()))
	//self.timers.start()
	//self.seckill_url[self.sku_id] = self.get_seckill_url()
	//logger.info('访问商品的抢购连接...')
	//headers = {
	//	'User-Agent': self.user_agent,
	//		'Host': 'marathon.jd.com',
	//		'Referer': 'https://item.jd.com/{}.html'.format(self.sku_id),
	//}
	//self.session.get(
	//	url=self.seckill_url.get(
	//	self.sku_id),
	//	headers=headers,
	//	allow_redirects=False)
	getSkuTittle()
	getUserName()
	getSecKillURL()
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

	url := `https://itemko.jd.com/itemShowBtn`

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		Sugar.Error(err)
	}

	req.Header.Set("User-Agent", loginInfo.userAgent)
	req.Header.Set("Referer", "https://item.jd.com/100012043978.html")
	req.Header.Set("Host", "itemko.jd.com")

	query := req.URL.Query()

	rand.Seed(time.Now().Unix())
	callback := "jQuery" + strconv.Itoa(rand.Intn(8999998)+1000000)
	query.Add("callback", callback)
	query.Add("skuId", "100012043978")
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

	fmt.Println(resp.StatusCode)
	fmt.Println(string(body))

	r := regexp.MustCompile(`"url":"([^"]+)"`)
	if r.Match(body) {
		loginInfo.secKillURL = "https:" + r.FindStringSubmatch(string(body))[1]
		Sugar.Infof("seckill URL:%v", loginInfo.secKillURL)
	} else {
		Sugar.Fatal("get second kill URL failure")
	}

}

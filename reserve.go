package main

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

func reserve() {

	checkLogin()
	getSkuTittle()
	getReserveURL()
	makeReserve()

}

func getReserveURL() {
	url := `https://yushou.jd.com/youshouinfo.action`

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		Sugar.Error(err)
	}

	req.Header.Set("User-Agent", secKillInfo.basic.userAgent)
	req.Header.Set("Referer", "https://item.jd.com/100012043978.html")

	query := req.URL.Query()
	query.Add("callback", "fetchJSON")
	query.Add("sku", secKillInfo.basic.skuID)
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

	r := regexp.MustCompile(`"url":"([^"]*)"`)
	if r.Match(body) {
		secKillInfo.reserver.URL = "https:" + r.FindStringSubmatch(string(body))[1]
	} else {
		Sugar.Fatal("reserve failure")
	}

}

func getSkuTittle() {

	//"""获取商品名称"""

	url := "https://item.jd.com/100012043978.html"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	resp, err := secKillInfo.basic.client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	r := regexp.MustCompile(`<title>(.*)</title>`)
	if r.Match(body) {
		Sugar.Infof("good title: %v", r.FindStringSubmatch(string(body))[1])
	}

}

func makeReserve() {

	req, err := http.NewRequest(http.MethodGet, secKillInfo.reserver.URL, nil)
	if err != nil {
		Sugar.Fatal(err)
	}

	query := req.URL.Query()
	query.Add("shopId", "1000085463")
	query.Add("isPlusLimit", "1")
	req.URL.RawQuery = query.Encode()

	resp, err := secKillInfo.basic.client.Do(req)
	if err != nil {
		Sugar.Fatal(err)
	}

	defer resp.Body.Close()

	Sugar.Info("reserve successful")
}

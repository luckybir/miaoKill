package main

import (
	"github.com/tidwall/gjson"
	"github.com/zellyn/kooky"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"time"
)



func checkLogin() {

	getLoginStatus()

	if loginInfo.isLogin == false {
		getLoginPage()
		getLoginQRCode()

		for i := 0; i < 10; i++ {
			time.Sleep(5 * time.Second)

			getQRCodeTicket()
			if loginInfo.ticket != "" {
				break
			}
		}

		if loginInfo.ticket == "" {
			Sugar.Fatal("fail to get login token")
		}

		validateQRCodeTicket()
	}

	getLoginStatus()
	if loginInfo.isLogin {
		Sugar.Info("login success")
	} else {
		Sugar.Fatal("login failure")
	}

}

func getLoginStatus() {
	url := "https://order.jd.com/center/list.action"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		Sugar.Fatal(err)
	}

	req.Header.Set("User-Agent", loginInfo.userAgent)

	resp, err := loginInfo.client.Do(req)
	if err != nil {
		Sugar.Fatal(err)
	}

	defer resp.Body.Close()

	//respBodyUTF8 := transform.NewReader(resp.Body, simplifiedchinese.GBK.NewDecoder())
	//body, err := ioutil.ReadAll(respBodyUTF8)
	//if err != nil {
	//	Sugar.Fatal(err)
	//}

	if resp.StatusCode == 200 {
		loginInfo.isLogin = true
	}

	//Sugar.Info(resp.StatusCode)
	//fmt.Println(string(body))
	//
	//
	//fmt.Printf("cookie: %+v\n",req.Cookies())
	//fmt.Printf("cookie2:%+v\n",loginInfo.client.Jar)

}

func getCookie() {
	cookies = make([]*http.Cookie, 0, 0)
	browserCookies := kooky.ReadCookies(kooky.Valid, kooky.DomainContains("jd"))
	for _, browserCookie := range browserCookies {
		cookie := &http.Cookie{Name: browserCookie.Name, Value: browserCookie.Value, HttpOnly: browserCookie.HttpOnly}
		cookies = append(cookies, cookie)
	}

	//Sugar.Infof("cookies length:%v", len(cookies))
}

func getLoginPage() {

	req, err := http.NewRequest("GET", "https://passport.jd.com/new/login.aspx", nil)
	if err != nil {
		Sugar.Fatal(err)
	}

	req.Header.Set("User-Agent", loginInfo.userAgent)
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3")

	resp, err := loginInfo.client.Do(req)
	if err != nil {
		Sugar.Fatal(err)
	}

	defer resp.Body.Close()

	Sugar.Infof("getLoginPage status %v", resp.StatusCode)

}

func getLoginQRCode() {
	req, err := http.NewRequest("GET", "https://qr.m.jd.com/show", nil)
	if err != nil {
		Sugar.Fatal(err)
	}

	query := req.URL.Query()
	query.Add("appid", "133")
	query.Add("size", "147")
	query.Add("t", "")
	req.URL.RawQuery = query.Encode()

	req.Header.Set("User-Agent", loginInfo.userAgent)
	req.Header.Set("Referer", "https://passport.jd.com/new/login.aspx")
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")

	resp, err := loginInfo.client.Do(req)
	if err != nil {
		Sugar.Fatal(err)
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		Sugar.Fatal("QRCode err")
	}

	err = ioutil.WriteFile("./QRCode.png", body, 0666)
	if err != nil {
		Sugar.Fatal(err)
	}

	Sugar.Info("scan QRCode")

	loginInfo.cookie = resp.Cookies()

}

func getQRCodeTicket() {
	req, err := http.NewRequest("GET", "https://qr.m.jd.com/check", nil)
	if err != nil {
		Sugar.Fatal(err)
	}

	query := req.URL.Query()
	query.Add("appid", "133")

	rand.Seed(time.Now().Unix())
	callback := "jQuery" + strconv.Itoa(rand.Intn(8999998)+1000000)
	query.Add("callback", callback)

	for _, cookie := range loginInfo.cookie {
		if cookie.Name == "wlfstk_smdl" {
			query.Add("token", cookie.Value)
		}

		req.AddCookie(&http.Cookie{Name: cookie.Name, Value: cookie.Value, Domain: cookie.Domain, Path: cookie.Path})
	}

	query.Add("_", string(time.Now().Unix()*1000))
	req.URL.RawQuery = query.Encode()

	req.Header.Set("User-Agent", loginInfo.userAgent)
	req.Header.Set("Referer", "https://passport.jd.com/new/login.aspx")

	resp, err := loginInfo.client.Do(req)
	if err != nil {
		Sugar.Fatal(err)
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode == 200 {
		r := regexp.MustCompile(`"code" : 200,`)
		if r.Match(body) {

			r = regexp.MustCompile(`"ticket" : "(.*)"`)
			if r.Match(body) {
				loginInfo.ticket = r.FindStringSubmatch(string(body))[1]
			}

		}
	}

}

func validateQRCodeTicket() {

	req, err := http.NewRequest("GET", "https://passport.jd.com/uc/qrCodeTicketValidation", nil)
	if err != nil {
		Sugar.Fatal(err)
	}

	query := req.URL.Query()
	query.Add("t", loginInfo.ticket)
	req.URL.RawQuery = query.Encode()

	req.Header.Set("User-Agent", loginInfo.userAgent)
	req.Header.Set("Referer", "https://passport.jd.com/uc/login?ltype=logout")

	resp, err := loginInfo.client.Do(req)
	if err != nil {
		Sugar.Fatal(err)
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	returnCode := gjson.GetBytes(body, "returnCode").Int()
	if returnCode != 0 {
		Sugar.Fatal("validate QRCode fail")
	} else {
		Sugar.Info("validate QRCode successful")
	}

}

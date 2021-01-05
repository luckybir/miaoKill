package main

import (
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"
)



func checkLogin() {

	getLoginStatus()

	if secKillInfo.login.isLogin == false {
		getLoginPage()
		getLoginQRCode()

		for i := 0; i < 10; i++ {
			time.Sleep(5 * time.Second)

			getQRCodeTicket()
			if secKillInfo.login.ticket != "" {
				break
			}
		}

		if secKillInfo.login.ticket == "" {
			Sugar.Fatal("fail to get login token")
		}

		validateQRCodeTicket()
	}

	getLoginStatus()
	if secKillInfo.login.isLogin {
		Sugar.Info("login")
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

	req.Header.Set("User-Agent", secKillInfo.basic.userAgent)

	resp, err := secKillInfo.basic.client.Do(req)
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
		secKillInfo.login.isLogin = true
	}

	//Sugar.Info(resp.StatusCode)
	//fmt.Println(string(body))
	//
	//
	//fmt.Printf("cookie: %+v\n",req.Cookies())
	//fmt.Printf("cookie2:%+v\n",secKillInfo.basic.client.Jar)

}



func getLoginPage() {

	req, err := http.NewRequest("GET", "https://passport.jd.com/new/login.aspx", nil)
	if err != nil {
		Sugar.Fatal(err)
	}

	req.Header.Set("User-Agent", secKillInfo.basic.userAgent)
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3")

	resp, err := secKillInfo.basic.client.Do(req)
	if err != nil {
		Sugar.Fatal(err)
	}

	defer resp.Body.Close()

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

	req.Header.Set("User-Agent", secKillInfo.basic.userAgent)
	req.Header.Set("Referer", "https://passport.jd.com/new/login.aspx")
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")

	resp, err := secKillInfo.basic.client.Do(req)
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

	cookies:= resp.Cookies()
	for _,cookie := range cookies{
		if cookie.Name == "wlfstk_smdl"{
			secKillInfo.login.token = cookie.Value
		}
	}

	Sugar.Info("scan QRCode immediately...")
}

func getQRCodeTicket() {
	req, err := http.NewRequest("GET", "https://qr.m.jd.com/check", nil)
	if err != nil {
		Sugar.Fatal(err)
	}

	query := req.URL.Query()
	query.Add("appid", "133")

	callback := "jQuery" + strconv.Itoa(randomInt(1000000,9999999))
	query.Add("callback", callback)
	query.Add("token", secKillInfo.login.token)

	query.Add("_", string(time.Now().Unix()*1000))
	req.URL.RawQuery = query.Encode()

	req.Header.Set("User-Agent", secKillInfo.basic.userAgent)
	req.Header.Set("Referer", "https://passport.jd.com/new/login.aspx")

	resp, err := secKillInfo.basic.client.Do(req)
	if err != nil {
		Sugar.Fatal(err)
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode == 200 {
		r := regexp.MustCompile(`"code" : 200,`)
		if r.Match(body) {

			r = regexp.MustCompile(`"ticket" : "([^"]*)"`)
			if r.Match(body) {
				secKillInfo.login.ticket = r.FindStringSubmatch(string(body))[1]
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
	query.Add("t", secKillInfo.login.ticket)
	req.URL.RawQuery = query.Encode()

	req.Header.Set("User-Agent", secKillInfo.basic.userAgent)
	req.Header.Set("Referer", "https://passport.jd.com/uc/login?ltype=logout")

	resp, err := secKillInfo.basic.client.Do(req)
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

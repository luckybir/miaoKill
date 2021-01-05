package main

import (
	_ "github.com/zellyn/kooky/allbrowsers" // register cookie store finders!
	"go.uber.org/zap"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"time"
)

type QRCodeLoginInfo struct {
	client           *http.Client
	userAgent        string
	cookie           []*http.Cookie
	ticket           string
	isLogin          bool
	reserveURL       string
	secKillURL       string
	serverTimeOffset int64
	skuID string
}

var Sugar *zap.SugaredLogger
var cookies []*http.Cookie

var loginInfo QRCodeLoginInfo

func init() {
	Sugar = zap.NewExample().Sugar()
	defer Sugar.Sync()

	loginInfo.userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.36"

	loginInfo.client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	jar, _ := cookiejar.New(nil)
	loginInfo.client.Jar = jar

	loginInfo.skuID = "100012043978"

	getJdTimeOffset()

	rand.Seed(time.Now().Unix())

}

// https://github.com/huanghyw/jd_seckill/tree/master

func main() {
	//tttt()
	//reserve()
	secondKill()
}

func tttt() {


	Sugar.Fatal("return")
}

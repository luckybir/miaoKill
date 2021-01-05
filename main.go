package main

import (
	_ "github.com/zellyn/kooky/allbrowsers" // register cookie store finders!
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"time"
)

type Config struct {
	Eid       string `yaml:"eid"`
	Fp        string `yaml:"fp"`
	SkuID     string `yaml:"sku_id"`
	UserAgent string `yaml:"user_agent"`
}

type jdSecondKillInfo struct {
	basic struct {
		client           *http.Client
		userAgent        string
		serverTimeOffset int64
		skuID            string
		eid              string
		fp               string
	}

	login struct {
		ticket  string
		isLogin bool
		token   string
	}

	reserver struct {
		URL string
	}

	secKill struct {
		URL string
	}
}

var Sugar *zap.SugaredLogger

var secKillInfo jdSecondKillInfo

func init() {
	Sugar = zap.NewExample().Sugar()
	defer Sugar.Sync()

	initsecKillInfo()

	getJdTimeOffset()

	rand.Seed(time.Now().UnixNano())

}

// https://github.com/huanghyw/jd_seckill/tree/master

func main() {
	//tttt()
	reserve()
	secondKill()
}

func tttt() {
	Sugar.Fatal("return")
}

func initsecKillInfo() {

	configText, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		Sugar.Fatal(err)
	}

	config := Config{}

	err = yaml.Unmarshal(configText, &config)
	if err != nil {
		Sugar.Fatal(err)
	}

	secKillInfo.basic.userAgent = config.UserAgent

	// collect cookies from response
	jar, _ := cookiejar.New(nil)
	secKillInfo.basic.client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Jar: jar,
	}

	secKillInfo.basic.skuID = config.SkuID
	secKillInfo.basic.eid = config.Eid
	secKillInfo.basic.fp = config.Fp
}

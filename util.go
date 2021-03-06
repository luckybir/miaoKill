package main

import (
	"fmt"
	"github.com/tidwall/gjson"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

func printGBKBody(respGBKbody io.ReadCloser) {
	respBodyUTF8 := transform.NewReader(respGBKbody, simplifiedchinese.GBK.NewDecoder())
	body, err := ioutil.ReadAll(respBodyUTF8)
	if err != nil {
		Sugar.Fatal(err)
	}

	fmt.Println(string(body))
}

func getJdTimeOffset(){
	resp,err := http.Get("https://a.jd.com//ajax/queryServerData.html")
	if err!=nil {
		Sugar.Fatal(err)
	}

	defer resp.Body.Close()

	body,err := ioutil.ReadAll(resp.Body)
	if err!=nil {
		Sugar.Fatal(err)
	}

	localTime := time.Now().Unix() * 1000
	serverTime := gjson.GetBytes(body,"serverTime").Int()

	secKillInfo.basic.serverTimeOffset = serverTime - localTime

	Sugar.Infof("本地时间与京东服务器时间误差为%v毫秒",secKillInfo.basic.serverTimeOffset)
}

func waitRandomTime(){
	r:= rand.Intn(300) + 100

	time.Sleep(time.Duration(r) * time.Microsecond)
}

func randomInt(min int ,max int)int{

	return rand.Intn(max - min) + min
}
package main

import (
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
)

func printGBKBody(respGBKbody io.ReadCloser) {
	respBodyUTF8 := transform.NewReader(respGBKbody, simplifiedchinese.GBK.NewDecoder())
	body, err := ioutil.ReadAll(respBodyUTF8)
	if err != nil {
		Sugar.Fatal(err)
	}

	fmt.Println(string(body))


}
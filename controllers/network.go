package controllers

import (
	"fmt"
	"net/http"
	"strings"
)

var network = map[int64]string{
	9200:  "https://rpc.tokchain.org",    // TOK
	15042: "https://devrpc.tokchain.org", // TOK-DEV
	1:     "https://mainnet.infura.io/v3/38c9c7fadd854c8d8ea3779728c11937",
	56:    "https://burned-clean-crater.bsc.quiknode.pro/5cacde470dbfb4087eab934b54e6b05fc59e1e92",
}

var chainId int64 = 9200
var Key = ""
var multicallAddr = "0xC1729FF538d353Ce0d477FdB5EbEF1000eeEc8eb"
var usadAddr = "0xdcd23789633479A052B881d90556EC5957324C50"
var consolidationAddr = "0xefdf028c872d2b66abe7b6ce749604f17e9c7721"

func PostLark(message string) {
	url := "https://open.larksuite.com/open-apis/bot/v2/hook/36802ff3-5f87-4769-810d-40fd9db9b775"
	payload := fmt.Sprintf(`{
    	"msg_type": "text",
    	"content": {
        "text": "%s"
    		}
	}`, message)

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, strings.NewReader(payload))

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

}

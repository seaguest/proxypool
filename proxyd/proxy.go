package main

import (
	"time"

	request "github.com/imroc/req"
	"github.com/seaguest/log"
)

const (
	timeout = 10000
)

func main() {
	req := request.New()
	req.SetTimeout(time.Duration(timeout) * time.Millisecond)
	req.SetProxyUrl("http://localhost:8080")

	headers := request.Header{
		"Content-Type": "application/json; charset=UTF-8",
		"User-Agent":   "Dalvik/1.6.0 (Linux; U; Android 4.4.2; HTC D816t Build/KOT49H)",
	}

	requestUrl := "http://localhost:9001/ping"
	resp, err := req.Get(requestUrl, headers)
	if err != nil {
		log.Error(err)
		return
	}

	type response struct {
		Anonymity int `json:"anonymity"`
		ErrCode   int `json:"err_code"`
	}

	var output response
	if err := resp.ToJSON(&output); err != nil {
		log.Error(err)
		return
	}

	log.Error(output)
}

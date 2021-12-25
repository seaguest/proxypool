package main

import (
	"time"

	"github.com/seaguest/log"
	"github.com/seaguest/proxypool/proxy"
)

type response struct {
	ErrCode   int `json:"err_code"`
	Anonymity int `json:"anonymity"`
}

func main() {
	px := proxy.GetGeneralProxy("127.0.0.1:6379", "", "test", valid)
	if px == nil {
		log.Error("empty proxy....")
		return
	}

	url := "http://123.56.144.9:9001/ping"

	for i := 0; i < 5000; i++ {
		go func() {
			for {
				var r response
				if err := px.Get(url, nil, &r, 10000); err != nil {
					log.Error(err)
					continue
				}
				log.Error("-----------received response---------------", r)
				time.Sleep(time.Second * 1)
			}
		}()
	}

	for {
		time.Sleep(time.Minute * 1)
	}
}

func valid(r interface{}) bool {
	resp := r.(*response)
	if resp.ErrCode == 0 {
		return true
	}
	return false
}

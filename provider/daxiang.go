package provider

import (
	"fmt"

	"github.com/imroc/req"
	"github.com/seaguest/log"
)

/*
{
  "msg": "",
  "code": 0,
  "data": {
    "count": 10,
    "proxy_list": [
      "124.172.117.189:80",
      "219.133.31.120:8888",
      "183.237.194.145:8080",
      "183.62.172.50:9999",
      "163.125.157.243:8888",
      "183.57.42.79:81",
      "202.103.150.70:8088",
      "182.254.129.124:80",
      "58.251.132.181:8888",
      "112.95.241.76:80"
    ]
  }
}

*/

const (
	daxiangApiUrl = "http://s4.proxypool.mayidaili.com:8123"
)

type DaxiangProxyProvider struct {
	// nothing to define
}

type daxiangResponse struct {
	Msg  string `json:"msg"`
	Code int    `json:"code"`
	Data struct {
		count int      `json:"code"`
		List  []string `json:"proxy_list"`
	} `json:"data"`
}

func (p *DaxiangProxyProvider) FetchProxy() ([]string, error) {
	resp, err := req.Get(daxiangApiUrl)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var response daxiangResponse
	if err := resp.ToJSON(&response); err != nil {
		log.Error(err)
		return nil, err
	}

	if response.Code != 0 {
		err := fmt.Errorf("error returned [%d]: [%s]", response.Code, response.Msg)
		log.Error(err)
		return nil, err
	}

	if response.Data.count == 0 {
		err := fmt.Errorf("empty proxy_list [%d]", response.Data.count)
		log.Error(err)
		return nil, err
	}

	return response.Data.List, nil
}

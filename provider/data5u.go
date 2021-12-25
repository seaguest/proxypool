package provider

import (
	"fmt"

	"github.com/imroc/req"
	"github.com/seaguest/log"
)

/*
{
	success: true,
	msg: '',
	data:[
		{
	    	ip: '19.203.43.22',
	    	port: 8089,
	    	country: '中国',
	    	province: '浙江省',
	    	city: '杭州市',
	    	isp: '电信',
	    	type: 'http',
	    	anonymity: 1,
	    	connectTimeMs: 36
	    },
	    {
	    	ip: '202.43.143.156',
	    	port: 80,
	    	country: '中国',
	    	province: '北京市',
	    	city: '北京市',
	    	isp: '铁通',
	    	type: 'http,https',
	    	anonymity: 2,
	    	connectTimeMs: 360
	    }
	]
}

*/

const (
	data5uApiUrl = "http://api.ip.data5u.com/api/get.shtml?order=cc6da89b4aff134b75e31ec0cf251bca&num=9999&area=%E4%B8%AD%E5%9B%BD&carrier=0&protocol=0&an1=1&an2=2&an3=3&sp1=1&sp2=2&sp3=3&sort=1&distinct=0&rettype=0&seprator=%0D%0A"
)

type Data5uProxyProvider struct {
	// nothing to define
}

type data5uResponse struct {
	Msg     string `json:"msg"`
	Success bool   `json:"success"`
	Data    []*struct {
		Ip   string `json:"ip"`
		Port int    `json:"port"`
	} `json:"data"`
}

func (p *Data5uProxyProvider) FetchProxy() ([]string, error) {
	resp, err := req.Get(data5uApiUrl)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var response data5uResponse
	if err := resp.ToJSON(&response); err != nil {
		log.Error(err)
		return nil, err
	}

	if !response.Success {
		err := fmt.Errorf("error returned [%t]: [%s]", response.Success, response.Msg)
		log.Error(err)
		return nil, err
	}

	if len(response.Data) == 0 {
		err := fmt.Errorf("empty proxy_list [%d]", len(response.Data))
		log.Error(err)
		return nil, err
	}

	log.Error("-------------------------------------------", len(response.Data))

	var proxies []string
	for _, data5uProxy := range response.Data {
		proxy := fmt.Sprintf("%s:%d", data5uProxy.Ip, data5uProxy.Port)
		proxies = append(proxies, proxy)
	}
	return proxies, nil
}

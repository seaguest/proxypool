package proxypool

import (
	"fmt"
	"time"

	request "github.com/imroc/req"
	"github.com/seaguest/log"
)

// check if a proxy is availale, return the rtt, anonymity
func validateProxy(ip, port string) (int, int, bool) {
	proxyUrl := fmt.Sprintf("http://%s:%s", ip, port)

	start := time.Now()

	req := request.New()
	req.SetTimeout(validationTimeout)
	// set proxy
	req.SetProxyUrl(proxyUrl)

	input := make(map[string]interface{})
	input["ip"] = ip

	resp, err := req.Get(validationUrl, request.QueryParam(input))
	if err != nil {
		log.Error(err)
		return 0, 0, false
	}

	type response struct {
		ErrCode   int `json:"err_code"`
		Anonymity int `json:"anonymity"`
	}

	var r response
	if err = resp.ToJSON(&r); err != nil {
		// log.Error(err)
		return 0, 0, false
	}

	// log.Error("--------------validation response------------", r)

	rtt := time.Since(start) / time.Millisecond
	return int(rtt), r.Anonymity, true
}

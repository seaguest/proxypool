package proxy

import (
	"time"

	request "github.com/imroc/req"
	"github.com/seaguest/log"
	"github.com/seaguest/proxypool"
)

const (
	abuyunProxyServer = "proxypool.abuyun.com:9020"
)

type AbuyunProxy struct {
	License   string
	SecretKey string
}

func GetAbuyunProxy(license, secretKey string) *AbuyunProxy {
	proxy := &AbuyunProxy{}
	proxy.License = license
	proxy.SecretKey = secretKey
	return proxy
}

func (p *AbuyunProxy) Get(url string, input, output interface{}, timeout int) error {
	req := request.New()

	// add timeout
	if timeout == 0 {
		timeout = proxypool.DefaultTimeout
	}
	req.SetTimeout(time.Duration(timeout) * time.Millisecond)

	// set proxy
	proxyURL := "http://" + p.License + ":" + p.SecretKey + "@" + abuyunProxyServer
	req.SetProxyUrl(proxyURL)

	headers := request.Header{
		"Proxy-Switch-Ip": "yes",
		"Content-Type":    "application/json; charset=UTF-8",
		"User-Agent":      "Dalvik/1.6.0 (Linux; U; Android 4.4.2; HTC D816t Build/KOT49H)",
	}

	in := input.(map[string]interface{})
	resp, err := req.Get(url, headers, request.QueryParam(in))
	if err != nil {
		log.Error(err)
		return err
	}

	if err := resp.ToJSON(&output); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (p *AbuyunProxy) GetRaw(url string, input interface{}, timeout int) (string, error) {
	req := request.New()

	// add timeout
	if timeout == 0 {
		timeout = proxypool.DefaultTimeout
	}
	req.SetTimeout(time.Duration(timeout) * time.Millisecond)

	// set proxy
	proxyURL := "http://" + p.License + ":" + p.SecretKey + "@" + abuyunProxyServer
	req.SetProxyUrl(proxyURL)

	headers := request.Header{
		"Proxy-Switch-Ip": "yes",
		"Content-Type":    "application/json; charset=UTF-8",
		"User-Agent":      "Dalvik/1.6.0 (Linux; U; Android 4.4.2; HTC D816t Build/KOT49H)",
	}

	in := input.(map[string]interface{})
	resp, err := req.Get(url, headers, request.QueryParam(in))
	if err != nil {
		log.Error(err)
		return "", err
	}

	s, err := resp.ToString()
	if err != nil {
		log.Error(err)
		return "", err
	}

	return s, nil
}

func (p *AbuyunProxy) Post(url string, input, output interface{}, timeout int) error {
	req := request.New()

	// add timeout
	if timeout == 0 {
		timeout = proxypool.DefaultTimeout
	}
	req.SetTimeout(time.Duration(timeout) * time.Millisecond)

	// set proxy
	proxyURL := "http://" + p.License + ":" + p.SecretKey + "@" + abuyunProxyServer
	req.SetProxyUrl(proxyURL)

	headers := request.Header{
		"Proxy-Switch-Ip": "yes",
		"Content-Type":    "application/json; charset=UTF-8",
		"User-Agent":      "Dalvik/1.6.0 (Linux; U; Android 4.4.2; HTC D816t Build/KOT49H)",
	}

	resp, err := req.Post(url, headers, request.BodyJSON(input))
	if err != nil {
		log.Error(err)
		return err
	}

	if err := resp.ToJSON(&output); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

package proxy

import (
	"sort"
	"time"

	request "github.com/imroc/req"
	"github.com/seaguest/common/util"
	"github.com/seaguest/log"
	"github.com/seaguest/proxypool"
)

const (
	mayiProxyServer = "http://s4.proxy.mayidaili.com:8123"
)

type MayiProxy struct {
	AppKey    string
	AppSecret string
}

func GetMayiProxy(appKey, appSecret string) *MayiProxy {
	proxy := &MayiProxy{}
	proxy.AppKey = appKey
	proxy.AppSecret = appSecret
	return proxy
}

func (p *MayiProxy) Get(url string, input, output interface{}, timeout int) error {
	req := request.New()

	// add timeout
	if timeout == 0 {
		timeout = proxypool.DefaultTimeout
	}
	req.SetTimeout(time.Duration(timeout) * time.Millisecond)

	// set proxy
	req.SetProxyUrl(mayiProxyServer)

	headers := request.Header{
		"Proxy-Authorization": p.sign(),
		// "Content-Type":        "application/json; charset=UTF-8",
		// "User-Agent":          "Dalvik/1.6.0 (Linux; U; Android 4.4.2; HTC D816t Build/KOT49H)",
	}

	var inputParam request.QueryParam
	if input != nil {
		in := input.(map[string]string)
		inputParam = request.QueryParam(in)
	}

	resp, err := req.Get(url, headers, request.QueryParam(inputParam))
	if err != nil {
		log.Error(err)
		return err
	}

	//s, _ := resp.ToString()
	//log.Error("---------", s)

	if err := resp.ToJSON(&output); err != nil {
		s, _ := resp.ToString()
		log.Error("input-------------------------", input)
		log.Error("response-------------------------", s)
		log.Error(err)
		return err
	}
	return nil
}

func (p *MayiProxy) GetRaw(url string, input interface{}, timeout int) (string, error) {
	req := request.New()

	// add timeout
	if timeout == 0 {
		timeout = proxypool.DefaultTimeout
	}
	req.SetTimeout(time.Duration(timeout) * time.Millisecond)

	// set proxy
	req.SetProxyUrl(mayiProxyServer)

	headers := request.Header{
		"Proxy-Authorization": p.sign(),
		"Content-Type":        "application/json; charset=UTF-8",
		"User-Agent":          "Dalvik/1.6.0 (Linux; U; Android 4.4.2; HTC D816t Build/KOT49H)",
	}

	var inputParam request.QueryParam
	if input != nil {
		in := input.(map[string]interface{})
		inputParam = request.QueryParam(in)
	}

	resp, err := req.Get(url, headers, request.QueryParam(inputParam))
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

func (p *MayiProxy) Post(url string, input, output interface{}, timeout int) error {
	req := request.New()

	// add timeout
	if timeout == 0 {
		timeout = proxypool.DefaultTimeout
	}
	req.SetTimeout(time.Duration(timeout) * time.Millisecond)

	// set proxy
	req.SetProxyUrl(mayiProxyServer)

	headers := request.Header{
		"Proxy-Authorization": p.sign(),
		"Content-Type":        "application/json; charset=UTF-8",
		"User-Agent":          "Dalvik/1.6.0 (Linux; U; Android 4.4.2; HTC D816t Build/KOT49H)",
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

func (p *MayiProxy) sign() string {
	params := map[string]string{
		"app_key":   p.AppKey,
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	}
	codes := p.AppSecret
	auth := "MYH-AUTH-MD5 "
	keySort := []string{}
	for k, _ := range params {
		keySort = append(keySort, k)
	}
	sort.Strings(keySort)
	for _, v := range keySort {
		codes += v + params[v]
		auth += v + "=" + params[v] + "&"
	}
	codes += p.AppSecret
	auth += "sign=" + util.Md5Hash(codes)

	return auth
}

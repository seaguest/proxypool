package proxy

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	request "github.com/imroc/req"
	"github.com/seaguest/log"
	"github.com/seaguest/proxypool"
)

type GeneralProxy struct {
	channel string
	pool    *proxypool.ProxyPool
	valid   func(interface{}) bool
	Auth    map[string]string
}

func GetGeneralProxy(redisAddr, redisPassword, channel string, valid func(interface{}) bool) *GeneralProxy {
	proxy := &GeneralProxy{}
	proxy.channel = channel
	proxy.pool = proxypool.NewProxyPool(redisAddr, redisPassword, channel)
	proxy.valid = valid
	return proxy
}

func (p *GeneralProxy) SetAuth(auth map[string]string) {
	p.Auth = auth
}

func (p *GeneralProxy) Get(url string, input, output interface{}, timeout int) error {
	proxy := p.pool.Take()
	if proxy == nil {
		err := errors.New("no available proxy")
		log.Error(err)
		return err
	}
	proxyUrl := fmt.Sprintf("http://%s", proxy.Member)
	req := getClient(proxyUrl, timeout)

	headers := request.Header{
		"Content-Type": "application/json; charset=UTF-8",
		"User-Agent":   "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36",
		//"User-Agent":   "Dalvik/1.6.0 (Linux; U; Android 4.4.2; HTC D816t Build/KOT49H)",
	}

	var inputParam request.QueryParam
	if input != nil {
		in := input.(map[string]interface{})
		inputParam = request.QueryParam(in)
	}

	var resp *request.Resp
	var err error

	defer func() {
		if err != nil {
			p.pool.Delete(proxy.Member)
			return
		}

		// if no error occurs, free proxy
		p.pool.Free(proxy)
	}()

	resp, err = req.Get(url, headers, inputParam)
	if err != nil {
		// delete the proxy
		log.Error("proxy-------------------------", proxy)
		log.Error(err)
		return err
	}

	if resp.Response().StatusCode != http.StatusOK {
		// delete the proxy
		err = errors.New(fmt.Sprintf("bad response status [%s] !", resp.Response().Status))
		log.Error(err)
		return err
	}

	if err = resp.ToJSON(output); err != nil {
		log.Error(err)
		return err
	}

	if !p.valid(output) {
		s, _ := resp.ToString()
		log.Error("response-------------------------", s)
		err = errors.New("invalid response format")
		log.Error(err)
		return err
	}
	return nil
}

func (p *GeneralProxy) GetRaw(url string, input interface{}, timeout int) (string, error) {
	proxy := p.pool.Take()
	if proxy == nil {
		err := errors.New("no available proxy")
		log.Error(err)
		return "", err
	}
	proxyUrl := fmt.Sprintf("http://%s", proxy.Member)
	req := getClient(proxyUrl, timeout)

	headers := request.Header{
		"Content-Type": "application/json; charset=UTF-8",
		"User-Agent":   "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36",
		"Cookie":       fmt.Sprintf("token=%s", p.Auth["token"]),
	}

	var inputParam request.QueryParam
	if input != nil {
		in := input.(map[string]interface{})
		inputParam = request.QueryParam(in)
	}

	var resp *request.Resp
	var err error

	defer func() {
		if err != nil {
			p.pool.Delete(proxy.Member)
			return
		}

		// if no error occurs, free proxy
		p.pool.Free(proxy)
	}()

	resp, err = req.Get(url, headers, request.QueryParam(inputParam))
	if err != nil {
		log.Error(err)
		return "", err
	}

	if resp.Response().StatusCode != http.StatusOK {
		// delete the proxy
		err = errors.New(fmt.Sprintf("bad response status [%s] !", resp.Response().Status))
		log.Error(err)
		return "", err
	}

	s, err := resp.ToString()
	if err != nil {
		log.Error(err)
		return "", err
	}

	if !p.valid(s) {
		s, _ := resp.ToString()
		log.Error("response-------------------------", s)
		err = errors.New("invalid response format")
		log.Error(err)
		return "", err
	}

	return s, nil
}

func (p *GeneralProxy) Post(url string, input, output interface{}, timeout int) error {
	req := request.New()

	// add timeout
	if timeout == 0 {
		timeout = proxypool.DefaultTimeout
	}
	req.SetTimeout(time.Duration(timeout) * time.Millisecond)

	// set proxy
	proxy := p.pool.Take()
	if proxy == nil {
		err := errors.New("no available proxy")
		log.Error(err)
		return err
	}

	proxyUrl := fmt.Sprintf("http://%s", proxy.Member)
	req.SetProxyUrl(proxyUrl)

	headers := request.Header{
		"Content-Type": "application/json; charset=UTF-8",
		"User-Agent":   "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36",
		// "User-Agent":   "Dalvik/1.6.0 (Linux; U; Android 4.4.2; HTC D816t Build/KOT49H)",
	}

	var resp *request.Resp
	var err error

	defer func() {
		if err != nil {
			p.pool.Delete(proxy.Member)
			return
		}

		// if no error occurs, free proxy
		p.pool.Free(proxy)
	}()

	resp, err = req.Post(url, headers, request.BodyJSON(input))
	if err != nil {
		log.Error(err)
		return err
	}

	if resp.Response().StatusCode != http.StatusOK {
		err = errors.New(fmt.Sprintf("bad response status [%s] !", resp.Response().Status))
		log.Error(err)
		return err
	}

	if err = resp.ToJSON(&output); err != nil {
		log.Error(err)
		return err
	}

	if !p.valid(output) {
		s, _ := resp.ToString()
		log.Error("response-------------------------", s)
		err = errors.New("invalid response format")
		log.Error(err)
		return err
	}
	return nil
}

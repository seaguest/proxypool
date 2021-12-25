package proxy

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/seaguest/log"
	"github.com/seaguest/proxypool"
)

// interface that proxy must implements.
type Proxy interface {
	Get(url string, input, output interface{}, timeout int) error

	GetRaw(url string, input interface{}, timeout int) (string, error)

	Post(url string, input, output interface{}, timeout int) error

	SetAuth(auth map[string]string)
}

func do(method string, requestUrl string, headers map[string]string, input interface{}, proxyUrl string, timeout int) ([]byte, error) {
	proxy, err := url.Parse(proxyUrl)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	b, err := json.Marshal(input)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	req, _ := http.NewRequest(method, requestUrl, bytes.NewReader(b))
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if timeout == 0 {
		timeout = proxypool.DefaultTimeout
	}

	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(time.Duration(timeout) * time.Millisecond)
				c, err := net.DialTimeout(netw, addr, time.Duration(timeout)*time.Millisecond)
				if err != nil {
					log.Error(err)
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
			Proxy:           http.ProxyURL(proxy),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err = errors.New(fmt.Sprintf("bad response status [%s] !", resp.Status))
		log.Error(err)
		return nil, err
	}

	// close body read before return
	defer resp.Body.Close()
	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return out, nil
}

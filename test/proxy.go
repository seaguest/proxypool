package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	request "github.com/imroc/req"
	"github.com/seaguest/log"
	"github.com/seaguest/proxypool"
)

type response struct {
	ErrCode   int `json:"err_code"`
	Anonymity int `json:"anonymity"`
}

func main() {
	/*
		var r response
		err := sendRequest("http://123.56.144.9:9001/ping", "39.64.80.50", "8118", &r)
		log.Error(err)
		log.Error(r)

	*/

	url := "http://www.legalminer.com/ajax_search/get_html?&cause=AND%24xz&court=AND%24%E9%BB%91%E9%BE%99%E6%B1%9F%E7%9C%81%E7%BB%A5%E5%8C%96%E5%B8%82%E7%BB%A5%E6%A3%B1%E5%8E%BF%E4%BA%BA%E6%B0%91%E6%B3%95%E9%99%A2&page=2&searchType=cases&sortField=judgeDate&sortType=ASC&year=AND%242003"
	proxy := "141.196.64.231:8080"

	s, err := getRaw(url, proxy)
	if err != nil {
		log.Error(err)
		return
	}

	log.Error(s)
}

func sendRequest(requestUrl, ip, port string, output interface{}) error {
	req := request.New()

	req.SetTimeout(time.Duration(proxypool.DefaultTimeout) * time.Millisecond)

	proxyURL := fmt.Sprintf("http://%s:%s", ip, port)
	// set proxy
	req.SetProxyUrl(proxyURL)

	headers := request.Header{
		"Content-Type": "application/json; charset=UTF-8",
		"User-Agent":   "Dalvik/1.6.0 (Linux; U; Android 4.4.2; HTC D816t Build/KOT49H)",
	}

	input := make(map[string]string)
	input["ip"] = ip

	resp, err := req.Get(requestUrl, headers, request.BodyJSON(input))
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

func getRaw(url, proxy string) (string, error) {
	req := request.New()
	req.SetTimeout(1000 * time.Millisecond)

	proxyUrl := fmt.Sprintf("http://%s", proxy)
	req.SetProxyUrl(proxyUrl)

	headers := request.Header{
		"Content-Type": "application/json; charset=UTF-8",
		"User-Agent":   "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36",
		// "User-Agent":   "Dalvik/1.6.0 (Linux; U; Android 4.4.2; HTC D816t Build/KOT49H)",
	}

	resp, err := req.Get(url, headers)
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
	return s, err
}

package proxy

import (
	"time"

	request "github.com/imroc/req"
	"github.com/patrickmn/go-cache"
	"github.com/seaguest/proxypool"
)

var clientCache *cache.Cache

func init() {
	clientCache = cache.New(time.Minute*10, 0)
}

func getClient(proxy string, timeout int) *request.Req {
	clientItf, found := clientCache.Get(proxy)
	if found {
		return clientItf.(*request.Req)
	}

	c := request.New()
	// add timeout
	if timeout == 0 {
		timeout = proxypool.DefaultTimeout
	}
	c.SetTimeout(time.Duration(timeout) * time.Millisecond)
	c.SetProxyUrl(proxy)

	clientCache.SetDefault(proxy, c)
	return c
}

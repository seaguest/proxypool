package proxypool

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/patrickmn/go-cache"
	"github.com/seaguest/common/util"
	"github.com/seaguest/log"
)

const (
	defaultPoolBlockedCleanPeriod = time.Second * 60
	defaultValidationPeriod       = time.Second * 60
	blockCacheTTL                 = time.Second * 1
)

type ProxyPool struct {
	// redis pool
	pool *redis.Pool

	// channel name
	channel string

	// cache holding the channel blocked proxy
	blockCache *cache.Cache

	// mutex for blocked_proxy
	mu *sync.Mutex
}

// create proxy_pool for each channel
func NewProxyPool(redisAddr, redisPassword string, channel string) *ProxyPool {
	pp := &ProxyPool{}
	pp.pool = NewRedisPool(redisAddr, redisPassword)
	pp.channel = channel
	pp.blockCache = cache.New(blockCacheTTL, 0)
	pp.mu = new(sync.Mutex)

	// start blocked proxy clean service
	go pp.cleanBlockedProxy()

	// validate proxy in proxy pool
	go pp.validate()
	return pp
}

// take a proxy from proxy_pool, a proxy can be used only by one thread at the same time.
func (p *ProxyPool) Take() *Member {
	proxyPoolKey := getProxyPoolKey(p.channel)
	proxy, err := zpop(proxyPoolKey, p.pool)
	if err != nil {
		log.Error(err)
		return nil
	}

	// if no proxy available, reload
	if proxy == nil {
		p.mu.Lock()
		defer p.mu.Unlock()

		// find all existing proxies in proxy_center
		keys, err := getKeysByPattern(proxyPrefix+"*", p.pool)
		if err != nil {
			log.Error(err)
			return nil
		}

		for _, key := range keys {
			proxy := strings.TrimPrefix(key, proxyPrefix)
			// if proxy is in the blocked pool, then skip
			if p.isProxyBlocked(proxy) {
				continue
			}

			proxyPoolKey := getProxyPoolKey(p.channel)
			if err := zaddIncr(proxyPoolKey, proxy, 0, p.pool); err != nil {
				log.Error(err)
			}
		}

		// retry
		proxy, err = zpop(proxyPoolKey, p.pool)
		if err != nil {
			log.Error(err)
			return nil
		}
	}
	return proxy
}

// when the proxy is used, then return it to proxy_pool
func (p *ProxyPool) Free(proxy *Member) {
	proxyPoolKey := getProxyPoolKey(p.channel)
	if err := zaddIncr(proxyPoolKey, proxy.Member, int64(proxy.Score+1), p.pool); err != nil {
		log.Error(err)
	}
}

// when an ip is blocked , put it in blocked set AND delete it from pool, not return it,
func (p *ProxyPool) Delete(proxy string) {
	// add it to blocked pool
	proxyBlockedKey := getProxyBlockedKey(p.channel)
	ts := time.Now().Unix()

	if err := zadd(proxyBlockedKey, proxy, ts, p.pool); err != nil {
		log.Error(err)
	}

	// remove from pool
	proxyPoolKey := getProxyPoolKey(p.channel)
	if err := zrem(proxyPoolKey, proxy, p.pool); err != nil {
		log.Error(err)
	}
}

// if a proxy is in blocked set longer than specified time, then delete it.
func (p *ProxyPool) cleanBlockedProxy() {
	for {
		blockedProxies, err := p.getBlockedProxies()
		if err != nil {
			log.Error(err)
			return
		}

		for _, blockedProxy := range blockedProxies {
			if time.Now().Sub(time.Unix(int64(blockedProxy.Score), 0)) > defaultPoolBlockedCleanPeriod {
				// if blocked_proxy xpires, clean it
				proxyBlockedKey := getProxyBlockedKey(p.channel)
				zrem(proxyBlockedKey, blockedProxy.Member, p.pool)
			}
		}

		time.Sleep(defaultPoolBlockedCleanPeriod)
	}
}

// check if all proxies in proxypool exist in proxycenter
func (p *ProxyPool) validate() {
	for {
		proxies, err := p.getProxies()
		if err != nil {
			log.Error(err)
			return
		}

		// find all existing proxies in proxy_center
		allProxiesKeys, err := getKeysByPattern(proxyPrefix+"*", p.pool)
		if err != nil {
			log.Error(err)
			return
		}

		proxyBlockedKey := getProxyBlockedKey(p.channel)
		for _, proxy := range proxies {
			proxyKey := proxyPrefix + proxy.Member
			if !util.ContainString(allProxiesKeys, proxyKey) {
				// if proxy is not present in proxy center, then add it to blocked proxy
				ts := time.Now().Unix()

				if err := zadd(proxyBlockedKey, proxy.Member, ts, p.pool); err != nil {
					log.Error(err)
				}
			}
		}
		time.Sleep(defaultValidationPeriod)
	}
}

func (p *ProxyPool) getBlockProxyCacheKey() string {
	return fmt.Sprintf("blockedproxy_%s", p.channel)
}

func (p *ProxyPool) getBlockedProxies() ([]*Member, error) {
	key := p.getBlockProxyCacheKey()
	blockedIpItf, found := p.blockCache.Get(key)
	if found {
		// if found in cache, then return directly
		return blockedIpItf.([]*Member), nil
	}

	proxyBlockedKey := getProxyBlockedKey(p.channel)
	blockedProxies, err := zrange(proxyBlockedKey, p.pool)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	p.blockCache.SetDefault(key, blockedProxies)
	return blockedProxies, nil
}

func (p *ProxyPool) getProxyCacheKey() string {
	return fmt.Sprintf("proxy_%s", p.channel)
}

// no need lock here, no concurrent cal
func (p *ProxyPool) getProxies() ([]*Member, error) {
	key := p.getProxyCacheKey()
	proxyItf, found := p.blockCache.Get(key)
	if found {
		// if found in cache, then return directly
		return proxyItf.([]*Member), nil
	}

	proxyPoolKey := getProxyPoolKey(p.channel)
	proxies, err := zrange(proxyPoolKey, p.pool)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	p.blockCache.SetDefault(key, proxies)
	return proxies, nil
}

func (p *ProxyPool) isProxyBlocked(proxy string) bool {
	blockedProxies, err := p.getBlockedProxies()
	if err != nil {
		log.Error(err)
		return false
	}

	for _, blockedProxy := range blockedProxies {
		if blockedProxy.Member == proxy {
			return true
		}
	}
	return false
}

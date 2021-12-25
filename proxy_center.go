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
	"github.com/seaguest/proxypool/provider"
)

const (
	defaultCenterValidationPeriod = time.Second * 300 // in second
	defaultLoadPeriod             = time.Second * 2   // in seconds
	defaultBlockedCleanPeriod     = time.Second * 60  // in seconds
	defaultProxyCacheTTL          = time.Second * 1   // in seconds
	defaultBlockCacheTTL          = time.Second * 1   // in seconds
	defaultMaxRoutine             = 500
	defaultProxyCenterChanSize    = 2000
)

// ProxyCenter is responsible for fetching proxies from other sites, and managing the global proxy pool.
// it will validate the proxies in the global pool periodically.
type ProxyCenter struct {
	// redis pool
	pool *redis.Pool

	// proxy channel for validation
	proxyChan chan string

	// providers which fetch proxies from third sites
	providers []provider.ProxyProvider

	// proxy ttl in the proxy center, in second
	validationPeriod time.Duration

	// proxy load period, in second
	loadPeriod time.Duration

	// max routinue for validation
	maxRoutine int

	// cache holding the channel blocked proxy
	blockCache *cache.Cache

	// mutex for blocked_proxy
	mu sync.Mutex
}

// create a new *ProxyCenter
func NewProxyCenter(redisAddr, redisPassword string, validationPeriod, loadPeriod time.Duration, maxRoutine int) *ProxyCenter {
	p := &ProxyCenter{}
	p.pool = NewRedisPool(redisAddr, redisPassword)
	p.proxyChan = make(chan string, defaultProxyCenterChanSize)

	if validationPeriod == 0 {
		p.validationPeriod = defaultCenterValidationPeriod
	} else {
		p.validationPeriod = validationPeriod
	}

	if loadPeriod == 0 {
		p.loadPeriod = defaultLoadPeriod
	} else {
		p.loadPeriod = loadPeriod
	}

	if maxRoutine == 0 {
		p.maxRoutine = defaultMaxRoutine
	} else {
		p.maxRoutine = maxRoutine
	}

	p.blockCache = cache.New(defaultBlockCacheTTL, 0)

	// by default, add three predefined proxy_provider
	//p.addProvider(provider.New("data5u"))
	//p.addProvider(provider.New("daxiang"))
	p.addProvider(provider.New("kuai"))

	// start proxy fetching service
	go p.fetchProxy()

	// start the proxy validation service
	go p.validate()

	// load outdated proxy to validation queue
	go p.scan()

	// start the blocked proxy clean service
	go p.cleanBlockedProxy()
	return p
}

// add provider to proxy center
func (p *ProxyCenter) addProvider(provider provider.ProxyProvider) {
	p.providers = append(p.providers, provider)
}

// scan the proxies in proxy_center, enqueue for validation
func (p *ProxyCenter) scan() {
	for {
		keys, err := p.getAllProxyKeys()
		if err != nil {
			log.Error(err)
			return
		}

		var proxies []string
		for _, key := range keys {
			proxy, err := getProxy(key, p.pool)
			if err != nil {
				log.Error(err)
				continue
			}

			// only validate proxies which have been validated 5 minutes before
			if time.Now().Sub(time.Unix(proxy.ValidatedAt, 0)) > p.validationPeriod {
				proxies = append(proxies, fmt.Sprintf("%s:%s", proxy.Ip, proxy.Port))
			}
		}

		log.Error("-------------revalidate proxies...", len(proxies))

		// load proxies to channel to validate
		p.enqueue(proxies)

		time.Sleep(p.validationPeriod)
	}
}

// proxy validation service
func (p *ProxyCenter) validate() {
	processProxy := func(proxyStr string) {
		sps := strings.Split(proxyStr, ":")
		if len(sps) != 2 {
			return
		}

		ip := sps[0]
		port := sps[1]

		rtt, anonymity, valid := validateProxy(ip, port)
		if !valid {
			// if proxy is not valid, remove it from global pool.
			key := getProxyKey(ip, port)
			delKey(key, p.pool)

			// after remove the invalid proxy, add it to blocked proxy
			ts := time.Now().Unix()
			if err := zadd(proxyBlockedSet, proxyStr, ts, p.pool); err != nil {
				log.Error(err)
			}
		} else {
			// remove proxy from blocked
			zrem(proxyBlockedSet, proxyStr, p.pool)

			// save to proxy
			var proxy Proxy
			proxy.Ip = ip
			proxy.Port = port
			proxy.Rtt = rtt
			proxy.Anonymity = anonymity
			proxy.ValidatedAt = time.Now().Unix()
			saveProxy(&proxy, p.pool)
		}

	}

	for i := 0; i < p.maxRoutine; i++ {
		go func() {
			for {
				select {
				case proxy := <-p.proxyChan:
					processProxy(proxy)
				}
			}
		}()
	}
}

// add proxy to proxy_chan waiting for validation
func (p *ProxyCenter) enqueue(proxies []string) {
	for _, proxy := range proxies {
		select {
		case p.proxyChan <- proxy:
		}
	}
}

func (p *ProxyCenter) fetchProxy() {
	for _, pd := range p.providers {
		go func(pd provider.ProxyProvider) {
			ticker := time.NewTicker(p.loadPeriod)
			defer ticker.Stop()
			for {
				<-ticker.C
				proxies, err := pd.FetchProxy()
				if err != nil {
					log.Error(err)
					continue
				}

				// find all existing proxies in proxy_center
				keys, err := p.getAllProxyKeys()
				if err != nil {
					log.Error(err)
					return
				}

				var existingProxies []string
				for _, key := range keys {
					proxy := strings.TrimPrefix(key, proxyPrefix)
					existingProxies = append(existingProxies, proxy)
				}

				// find all blocked proxy
				blockedProxies, err := p.getBlockedProxies()
				if err != nil {
					log.Error(err)
					return
				}

				for _, blockedProxy := range blockedProxies {
					existingProxies = append(existingProxies, blockedProxy.Member)
				}
				_, _, addedProxies := util.IntersectString(existingProxies, proxies)

				log.Error("-------------new proxies...", len(addedProxies))
				p.enqueue(addedProxies)
			}
		}(pd)
	}
}

// if a proxy is in blocked set longer than specified time, then delete it.
func (p *ProxyCenter) cleanBlockedProxy() {
	ticker := time.NewTicker(defaultBlockedCleanPeriod)
	defer ticker.Stop()
	for {
		<-ticker.C

		blockedProxies, err := p.getBlockedProxies()
		if err != nil {
			log.Error(err)
			return
		}

		for _, blockedProxy := range blockedProxies {
			if time.Now().Sub(time.Unix(int64(blockedProxy.Score), 0)) > defaultBlockedCleanPeriod {
				// if blocked_proxy xpires, clean it
				zrem(proxyBlockedSet, blockedProxy.Member, p.pool)
			}
		}
	}
}

func (p *ProxyCenter) getBlockProxyCacheKey() string {
	return fmt.Sprint("blockedproxy")
}

func (p *ProxyCenter) getBlockedProxies() ([]*Member, error) {
	key := p.getBlockProxyCacheKey()
	blockedIpItf, found := p.blockCache.Get(key)
	if found {
		// if found in cache, then return directly
		return blockedIpItf.([]*Member), nil
	}

	blockedProxies, err := zrange(proxyBlockedSet, p.pool)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	p.blockCache.Set(key, blockedProxies, defaultBlockCacheTTL)
	return blockedProxies, nil
}

func (p *ProxyCenter) getAllProxyCacheKey() string {
	return fmt.Sprint("allproxy")
}

func (p *ProxyCenter) getAllProxyKeys() ([]string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := p.getAllProxyCacheKey()
	proxyItf, found := p.blockCache.Get(key)
	if found {
		// if found in cache, then return directly
		return proxyItf.([]string), nil
	}

	// find all existing proxies in proxy_center
	keys, err := getKeysByPattern(proxyPrefix+"*", p.pool)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	p.blockCache.Set(key, keys, defaultProxyCacheTTL)
	return keys, nil
}

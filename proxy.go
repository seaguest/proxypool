package proxypool

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
	"github.com/seaguest/log"
)

type Proxy struct {
	Ip          string `redis:"ip"`
	Port        string `redis:"port"`
	Anonymity   int    `redis:"anonymity"`
	Rtt         int    `redis:"rtt"`
	ValidatedAt int64  `redis:"validated_at"`
}

/**************** define the redis cache key ****************/
func getProxyKey(ip, port string) string {
	return fmt.Sprintf("%s%s:%s", proxyPrefix, ip, port)
}

func getProxyPoolKey(channel string) string {
	return fmt.Sprintf("%s%s", proxyPoolPrefix, channel)
}

func getProxyBlockedKey(channel string) string {
	return fmt.Sprintf("%s%s", proxyPoolBlockedPrefix, channel)
}

func saveProxy(p *Proxy, pool *redis.Pool) error {
	key := getProxyKey(p.Ip, p.Port)

	if err := setObject(key, p, pool); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func getProxy(key string, pool *redis.Pool) (*Proxy, error) {
	var proxy Proxy

	if err := getObject(key, &proxy, pool); err != nil {
		log.Error(err)
		return nil, err
	}
	return &proxy, nil
}

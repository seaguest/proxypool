package proxypool

import (
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/seaguest/log"
)

type Member struct {
	Member string `json:"member"`
	Score  int    `json:"score"`
}

func containsMember(members []*Member, proxy string) bool {
	for _, mem := range members {
		if mem.Member == proxy {
			return true
		}
	}
	return false
}

// 获取redis连接池对象
func NewRedisPool(address, password string) *redis.Pool {
	pool := &redis.Pool{
		MaxIdle:     50,
		IdleTimeout: 240 * time.Second,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		Dial: func() (redis.Conn, error) {
			return dial("tcp", address, password)
		},
	}
	return pool
}

func dial(network, address, password string) (redis.Conn, error) {
	c, err := redis.Dial(network, address)
	if err != nil {
		return nil, err
	}
	if password != "" {
		if _, err := c.Do("AUTH", password); err != nil {
			c.Close()
			return nil, err
		}
	}
	return c, err
}

func getKeysByPattern(pattern string, pool *redis.Pool) ([]string, error) {
	conn := pool.Get()
	defer conn.Close()

	if err := conn.Err(); err != nil {
		return []string{}, err
	}

	data, err := redis.Strings(conn.Do("KEYS", pattern))
	return data, err
}

func setObject(key string, value interface{}, pool *redis.Pool) error {
	conn := pool.Get()
	defer conn.Close()

	var err error
	if err := conn.Err(); err != nil {
		log.Error(err)
		return err
	}

	if _, err = conn.Do("HMSET", redis.Args{}.Add(key).AddFlat(value)...); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func getObject(key string, value interface{}, pool *redis.Pool) error {
	conn := pool.Get()
	defer conn.Close()

	var err error
	if err := conn.Err(); err != nil {
		log.Error(err)
		return err
	}

	v, err := redis.Values(conn.Do("HGETALL", key))
	if err != nil {
		log.Error(err)
		return err
	}

	if err := redis.ScanStruct(v, value); err != nil {
		log.Error(err)
		return err
	}
	return err
}

func zpop(key string, pool *redis.Pool) (*Member, error) {
	c := pool.Get()
	defer c.Close()

	var err error
	if err := c.Err(); err != nil {
		log.Error(err)
		return nil, err
	}

	var zpopScript = redis.NewScript(1, `
	    local r = redis.call('ZREVRANGE', KEYS[1], 0, 0, "WITHSCORES")
	    if r ~= nil then
	        local  k = r[1]
	        redis.call('ZREM', KEYS[1], k)
	    end
	    return r`)
	v, _ := redis.Strings(zpopScript.Do(c, key))
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// return nil if no record found
	if len(v) == 0 {
		return nil, nil
	}

	var proxy Member
	proxy.Member = v[0]
	proxy.Score, _ = strconv.Atoi(v[1])
	return &proxy, nil
}

func zrange(key string, pool *redis.Pool) ([]*Member, error) {
	c := pool.Get()
	defer c.Close()

	var err error
	if err := c.Err(); err != nil {
		log.Error(err)
		return nil, err
	}

	data, err := redis.Strings(c.Do("ZRANGE", key, 0, -1, "WITHSCORES"))

	var members []*Member
	for i := 0; i < len(data)/2; i++ {
		var member Member
		member.Member = data[2*i]
		member.Score, _ = strconv.Atoi(data[2*i+1])
		members = append(members, &member)
	}
	return members, err
}

func zadd(key, value string, score int64, pool *redis.Pool) error {
	c := pool.Get()
	defer c.Close()

	var err error
	if err := c.Err(); err != nil {
		log.Error(err)
		return err
	}

	_, err = c.Do("ZADD", key, score, value)
	return err
}

func zrem(key, member string, pool *redis.Pool) error {
	c := pool.Get()
	defer c.Close()

	var err error
	if err := c.Err(); err != nil {
		log.Error(err)
		return err
	}

	_, err = c.Do("ZREM", key, member)
	return err
}

func zaddIncr(key, value string, score int64, pool *redis.Pool) error {
	c := pool.Get()
	defer c.Close()

	var err error
	if err := c.Err(); err != nil {
		log.Error(err)
		return err
	}

	_, err = c.Do("ZADD", key, "INCR", score, value)
	return err
}

func zrank(key, value string, pool *redis.Pool) (interface{}, error) {
	c := pool.Get()
	defer c.Close()

	var err error
	if err := c.Err(); err != nil {
		log.Error(err)
		return nil, err
	}

	data, err := c.Do("ZRANK", key, value)
	return data, err
}

func zexists(key, value string, pool *redis.Pool) bool {
	val, err := zrank(key, value, pool)

	if err != nil {
		log.Error(err)
		return false
	}

	if val == nil {
		return false
	}
	return true
}

func exists(key string, pool *redis.Pool) bool {
	c := pool.Get()
	defer c.Close()

	var err error
	if err := c.Err(); err != nil {
		log.Error(err)
		return false
	}

	data, err := c.Do("EXISTS", key)
	if err != nil {
		log.Error(err)
		return false
	}

	val := data.(int64)
	if val == 0 {
		return false
	}
	return true
}

func delKey(key string, pool *redis.Pool) error {
	conn := pool.Get()
	defer conn.Close()

	var err error
	if err = conn.Err(); err != nil {
		return err
	}

	_, err = conn.Do("DEL", key)
	return err
}

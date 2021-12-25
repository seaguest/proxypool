package proxypool

import "time"

const (
	/****************** redis prefix setting ******************/
	proxyPrefix            = "proxy_"
	proxyBlockedSet        = "proxy_blocked"
	proxyPoolPrefix        = "proxypool_"
	proxyPoolBlockedPrefix = "proxypool_blocked_"

	/****************** proxy request timeout ******************/
	DefaultTimeout = 10000

	/****************** validation setting ******************/
	validationUrl     = "http://39.108.223.220:9001/ping"
	validationTimeout = time.Millisecond * 10000 // in milliseconds

	/****************** anonymity level ******************/
	AnonymityTransparent = 1
	AnonymityAnonymous   = 2
	AnonymityHigh        = 3
)

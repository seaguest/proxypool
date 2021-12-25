package provider

import "github.com/seaguest/log"

// a proxy provider implements a proxy fetch interface.
type ProxyProvider interface {

	// fetch the proxy list
	FetchProxy() ([]string, error)
}

func New(provider string) ProxyProvider {
	switch provider {
	case "data5u":
		return new(Data5uProxyProvider)
	case "daxiang":
		return new(DaxiangProxyProvider)
	case "kuai":
		return new(KuaiProxyProvider)
	default:
		log.Errorf("unknown provider [%s]", provider)
		return nil
	}

}

package main

import (
	"net/http"
	_ "net/http/pprof"
)

func main() {
	conf, err := LoadConfig("socks.config")
	if err != nil {
		ErrLog.Println("initGlobalConfig failed, err:", err)
		return
	}
	InfoLog.Println(conf)

	loadBalancer := NewLoadBalancer(conf.RemoteConfigs)
	dnsCache := NewDNSCache(conf.DNSCacheTimeout)
	upstreamConnector := NewUpstreamConnector(loadBalancer, dnsCache)

	go http.ListenAndServe(conf.PprofAddr, nil)

	httpProxy := NewHTTPProxy(upstreamConnector.ConnectUpstream)
	go httpProxy.Run(conf.HTTPProxyAddr)

	socks4Svr := NewSOCKS4Server(upstreamConnector.ConnectUpstream)
	go socks4Svr.Run(conf.SOCKS4Addr)

	socks5Svr := NewSocks5Server(conf.LocalCryptoMethod, conf.LocalCryptoPassword, upstreamConnector.ConnectUpstream)
	socks5Svr.Run(conf.SOCKS5Addr)
}

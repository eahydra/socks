package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
)

func main() {
	confGroup, err := LoadConfigGroup("socks.config")
	if err != nil {
		ErrLog.Println("initGlobalConfig failed, err:", err)
		return
	}
	InfoLog.Println(confGroup)

	for _, conf := range confGroup.AllConfig {
		loadBalancer := NewLoadBalancer(conf.AllUpstreamConfig)
		dnsCache := NewDNSCache(conf.DNSCacheTimeout)
		upstreamConnector := NewUpstreamConnector(loadBalancer, dnsCache)

		httpProxy := NewHTTPProxy(upstreamConnector.ConnectUpstream)
		go httpProxy.Run(conf.HTTPProxyAddr)

		socks4Svr := NewSOCKS4Server(upstreamConnector.ConnectUpstream)
		go socks4Svr.Run(conf.SOCKS4Addr)

		socks5Svr := NewSocks5Server(conf.LocalCryptoMethod, conf.LocalCryptoPassword, upstreamConnector.ConnectUpstream)
		go socks5Svr.Run(conf.SOCKS5Addr)
	}
	go http.ListenAndServe(confGroup.PprofAddr, nil)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Kill, os.Interrupt)
	<-sigChan
}

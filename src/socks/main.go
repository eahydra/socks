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

	go http.ListenAndServe(conf.PprofAddr, nil)

	httpProxy := NewHTTPProxy(loadBalancer)
	go httpProxy.Run(conf.HTTPProxyAddr)

	socks4Svr := NewSOCKS4Server(loadBalancer)
	go socks4Svr.Run(conf.SOCKS4Addr)

	socks5Svr := NewSocks5Server(conf.LocalCryptoMethod, []byte(conf.LocalCryptoPassword), loadBalancer)
	socks5Svr.Run(conf.SOCKS5Addr)
}

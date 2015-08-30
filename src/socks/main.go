package main

import (
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
)

func main() {
	confGroup, err := LoadConfigGroup("socks.config")
	if err != nil {
		ErrLog.Println("initGlobalConfig failed, err:", err)
		return
	}
	InfoLog.Println(confGroup)

	for _, conf := range confGroup.AllConfig {
		router := BuildUpstreamRouter(conf)
		runHTTPProxyServer(conf, router)
		runSOCKS4Server(conf, router)
		runSOCKS5Server(conf, router)
	}
	go http.ListenAndServe(confGroup.PprofAddr, nil)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Kill, os.Interrupt)
	<-sigChan
}

func BuildUpstreamRouter(conf Config) *UpstreamRouter {
	var routers []Router
	for _, upstreamConf := range conf.AllUpstreamConfig {
		var router Router
		switch strings.ToLower(upstreamConf.ServerType) {
		case "socks5":
			{
				socks5ClientFactory := func(conn net.Conn) SOCKClient {
					return NewSOCKS5Client(conn)
				}
				router = NewSOCKSRouter(upstreamConf.Addr, socks5ClientFactory,
					CipherConnDecorator(upstreamConf.CryptoMethod, upstreamConf.Password))
			}
		case "shadowsocks":
			{
				shadowSocksClientFactory := func(conn net.Conn) SOCKClient {
					return NewShadowSocksClient(conn)
				}
				router = NewSOCKSRouter(upstreamConf.Addr, shadowSocksClientFactory,
					CipherConnDecorator(upstreamConf.CryptoMethod, upstreamConf.Password))
			}
		default:
			{
				router = NewDirectRouter(conf.DNSCacheTimeout)
			}
		}
		routers = append(routers, router)
	}
	if len(routers) == 0 {
		router := NewDirectRouter(conf.DNSCacheTimeout)
		routers = append(routers, router)
	}
	return NewUpstreamRouter(NewUpstreamRouterBalancer(routers))
}

func runHTTPProxyServer(conf Config, router Router) {
	if conf.HTTPProxyAddr != "" {
		httpProxy := NewHTTPProxy(router)
		go httpProxy.Run(conf.HTTPProxyAddr)
	}
}

func runSOCKS4Server(conf Config, router Router) {
	if conf.SOCKS4Addr != "" {
		socks4Svr := NewSOCKS4Server(router)
		go socks4Svr.Run(conf.SOCKS4Addr)
	}
}

func runSOCKS5Server(conf Config, router Router) {
	if conf.SOCKS5Addr != "" {
		listener, err := net.Listen("tcp", conf.SOCKS5Addr)
		if err == nil {
			listener = NewDecorateListener(listener, CipherConnDecorator(conf.LocalCryptoMethod, conf.LocalCryptoPassword))
			socks5Svr := NewSocks5Server(router)
			go socks5Svr.Run(listener)
		}
	}
}

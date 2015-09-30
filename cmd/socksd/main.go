package main

import (
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"

	"github.com/eahydra/socks"
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
	var routers []socks.Router
	for _, upstreamConf := range conf.AllUpstreamConfig {
		var router socks.Router
		router = NewDirectRouter(conf.DNSCacheTimeout)
		switch strings.ToLower(upstreamConf.ServerType) {
		case "socks5":
			{
				clientFactory := func(conn net.Conn) SOCKClient {
					return socks.NewSOCKS5Client(conn)
				}

				router = NewSOCKSRouter(upstreamConf.Addr, router, clientFactory,
					CipherConnDecorator(upstreamConf.CryptoMethod, upstreamConf.Password))
			}
		case "shadowsocks":
			{
				clientFactory := func(conn net.Conn) SOCKClient {
					return socks.NewShadowSocksClient(conn)
				}
				router = NewSOCKSRouter(upstreamConf.Addr, router, clientFactory,
					CipherConnDecorator(upstreamConf.CryptoMethod, upstreamConf.Password))
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

func runHTTPProxyServer(conf Config, router socks.Router) {
	if conf.HTTPProxyAddr != "" {
		listener, err := net.Listen("tcp", conf.HTTPProxyAddr)
		if err != nil {
			return
		}
		go func() {
			defer listener.Close()
			httpProxy := socks.NewHTTPProxy(router)
			http.Serve(listener, httpProxy)
		}()
	}
}

func runSOCKS4Server(conf Config, router socks.Router) {
	if conf.SOCKS4Addr != "" {
		socks4Svr := socks.NewSOCKS4Server(router)
		go socks4Svr.Run(conf.SOCKS4Addr)
	}
}

func runSOCKS5Server(conf Config, router socks.Router) {
	if conf.SOCKS5Addr != "" {
		listener, err := net.Listen("tcp", conf.SOCKS5Addr)
		if err == nil {
			listener = NewDecorateListener(listener, CipherConnDecorator(conf.LocalCryptoMethod, conf.LocalCryptoPassword))
			socks5Svr := socks.NewSocks5Server(router)
			go socks5Svr.Run(listener)
		}
	}
}

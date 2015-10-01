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

func BuildUpstreamRouter(conf Config) socks.Dialer {
	var allForward []socks.Dialer
	for _, upstreamConf := range conf.AllUpstreamConfig {
		var forward socks.Dialer
		forward = NewDecorateDirect(conf.DNSCacheTimeout)
		cipherDecorator := NewCipherConnDecorator(upstreamConf.CryptoMethod, upstreamConf.Password)
		forward = NewDecorateClient(forward, cipherDecorator)

		var err error
		switch strings.ToLower(upstreamConf.ServerType) {
		case "socks5":
			{
				forward, err = socks.NewSocks5Client("tcp", upstreamConf.Addr, forward)
			}
		case "shadowsocks":
			{
				forward, err = socks.NewShadowSocksClient("tcp", upstreamConf.Addr, forward)
			}
		}
		if err != nil {
			ErrLog.Println("build upstream failed, err:", err, upstreamConf.ServerType, upstreamConf.Addr)
			continue
		}
		allForward = append(allForward, forward)
	}
	if len(allForward) == 0 {
		router := NewDecorateDirect(conf.DNSCacheTimeout)
		allForward = append(allForward, router)
	}
	return NewUpstreamDialer(allForward)
}

func runHTTPProxyServer(conf Config, router socks.Dialer) {
	if conf.HTTPProxyAddr != "" {
		listener, err := net.Listen("tcp", conf.HTTPProxyAddr)
		if err != nil {
			ErrLog.Println("net.Listen at ", conf.HTTPProxyAddr, " failed, err:", err)
			return
		}
		go func() {
			defer listener.Close()
			httpProxy := socks.NewHTTPProxy(router)
			http.Serve(listener, httpProxy)
		}()
	}
}

func runSOCKS4Server(conf Config, forward socks.Dialer) {
	if conf.SOCKS4Addr != "" {
		listener, err := net.Listen("tcp", conf.SOCKS4Addr)
		if err != nil {
			ErrLog.Println("net.Listen failed, err:", err, conf.SOCKS4Addr)
			return
		}
		cipherDecorator := NewCipherConnDecorator(conf.LocalCryptoMethod, conf.LocalCryptoPassword)
		listener = NewDecorateListener(listener, cipherDecorator)
		socks4Svr, err := socks.NewSocks4Server(forward)
		if err != nil {
			listener.Close()
			ErrLog.Println("socks.NewSocks4Server failed, err:", err)
		}
		go func() {
			defer listener.Close()
			socks4Svr.Serve(listener)
		}()
	}
}

func runSOCKS5Server(conf Config, forward socks.Dialer) {
	if conf.SOCKS5Addr != "" {
		listener, err := net.Listen("tcp", conf.SOCKS5Addr)
		if err != nil {
			ErrLog.Println("net.Listen failed, err:", err, conf.SOCKS5Addr)
			return
		}
		cipherDecorator := NewCipherConnDecorator(conf.LocalCryptoMethod, conf.LocalCryptoPassword)
		listener = NewDecorateListener(listener, cipherDecorator)
		socks5Svr, err := socks.NewSocks5Server(forward)
		if err != nil {
			listener.Close()
			ErrLog.Println("socks.NewSocks5Server failed, err:", err)
			return
		}
		go func() {
			defer listener.Close()
			socks5Svr.Serve(listener)
		}()
	}
}

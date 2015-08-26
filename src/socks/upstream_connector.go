package main

import (
	"net"
	"strings"
)

type ConnectUpstream func(addr string) (net.Conn, error)

type UpstreamConnector struct {
	loadBalancer LoadBalancer
	dnsCache     *DNSCache
}

func NewUpstreamConnector(loadBalancer LoadBalancer, dnsCache *DNSCache) *UpstreamConnector {
	return &UpstreamConnector{
		loadBalancer: loadBalancer,
		dnsCache:     dnsCache,
	}
}

func (u *UpstreamConnector) ConnectUpstream(addr string) (net.Conn, error) {
	serverType, proxyServer, cryptoMethod, password := u.loadBalancer()
	switch strings.ToLower(serverType) {
	default:
		fallthrough
	case "socks5":
		{
			if proxyServer != "" {
				socks5Client, err := DialSOCKS5(proxyServer, cryptoMethod, password)
				if err != nil {
					return nil, err
				}
				if err = socks5Client.ConnectUpstream(addr); err != nil {
					socks5Client.Close()
					return nil, err
				}
				return socks5Client, nil
			}
		}
	case "shadowsocks":
		{
			if proxyServer != "" {
				shadowSocksClient, err := DialShadowSocks(proxyServer, cryptoMethod, password)
				if err != nil {
					return nil, err
				}
				if err = shadowSocksClient.ConnectUpstream(addr); err != nil {
					shadowSocksClient.Close()
					return nil, err
				}
				return shadowSocksClient, nil
			}
		}
	}

	return u.DirectConnect(addr)
}

func (u *UpstreamConnector) DirectConnect(addr string) (net.Conn, error) {
	host, port, err := parseAddress(addr)
	if err != nil {
		return nil, err
	}
	var dest string
	var ipCached bool
	switch h := host.(type) {
	case net.IP:
		{
			dest = h.String()
			ipCached = true
		}
	case string:
		{
			if p, ok := u.dnsCache.Get(h); ok {
				dest = p.String()
				ipCached = true
			} else {
				dest = h
			}
		}
	}
	destConn, err := net.Dial("tcp", net.JoinHostPort(dest, port))
	if err != nil {
		return nil, err
	}
	if !ipCached {
		u.dnsCache.Set(host.(string), destConn.RemoteAddr().(*net.TCPAddr).IP)
	}
	return destConn, nil
}

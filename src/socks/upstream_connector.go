package main

import (
	"fmt"
	"net"
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
	upStreamServer, cryptoMethod, password := u.loadBalancer()
	if upStreamServer != "" {
		upStream, err := DialSOCKS5(upStreamServer, cryptoMethod, password)
		if err != nil {
			return nil, err
		}
		if err = upStream.ConnectUpstream(addr); err != nil {
			upStream.Close()
			return nil, err
		}
		return upStream, nil

	} else {
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
		destConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", dest, port))
		if err != nil {
			return nil, err
		}
		if !ipCached {
			u.dnsCache.Set(host.(string), destConn.RemoteAddr().(*net.TCPAddr).IP)
		}
		return destConn, nil
	}
}

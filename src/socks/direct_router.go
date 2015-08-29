package main

import "net"

type DirectRouter struct {
	dnsCache *DNSCache
}

func NewDirectRouter(dnsCacheTime int) *DirectRouter {
	return &DirectRouter{
		dnsCache: NewDNSCache(dnsCacheTime),
	}
}

func (d *DirectRouter) Do(address string) (net.Conn, error) {
	host, port, err := parseAddress(address)
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
			if p, ok := d.dnsCache.Get(h); ok {
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
		d.dnsCache.Set(host.(string), destConn.RemoteAddr().(*net.TCPAddr).IP)
	}
	return destConn, nil
}

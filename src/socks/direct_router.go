package main

import "net"

type DirectRouter struct {
	dnsCache *DNSCache
}

func NewDirectRouter(dnsCacheTime int) *DirectRouter {
	var dnsCache *DNSCache
	if dnsCacheTime != 0 {
		dnsCache = NewDNSCache(dnsCacheTime)
	}
	return &DirectRouter{
		dnsCache: dnsCache,
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
			dest = h
			if d.dnsCache != nil {
				if p, ok := d.dnsCache.Get(h); ok {
					dest = p.String()
					ipCached = true
				}
			}
		}
	}
	destConn, err := net.Dial("tcp", net.JoinHostPort(dest, port))
	if err != nil {
		return nil, err
	}
	if d.dnsCache != nil && !ipCached {
		d.dnsCache.Set(host.(string), destConn.RemoteAddr().(*net.TCPAddr).IP)
	}
	return destConn, nil
}

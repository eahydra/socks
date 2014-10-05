package main

import (
	"encoding/binary"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type HTTPProxyConn struct {
	*RemoteSocks
}

func (h *HTTPProxyConn) LocalAddr() net.Addr {
	return h.RemoteSocks.conn.LocalAddr()
}

func (h *HTTPProxyConn) RemoteAddr() net.Addr {
	return h.RemoteSocks.conn.RemoteAddr()
}

func (h *HTTPProxyConn) SetDeadline(t time.Time) error {
	return h.RemoteSocks.conn.SetDeadline(t)
}

func (h *HTTPProxyConn) SetReadDeadline(t time.Time) error {
	return h.RemoteSocks.conn.SetReadDeadline(t)
}

func (h *HTTPProxyConn) SetWriteDeadline(t time.Time) error {
	return h.RemoteSocks.conn.SetWriteDeadline(t)
}

type HTTPProxy struct {
	*httputil.ReverseProxy
	remoteServer       string
	remoteCryptoMethod string
	remotePassword     []byte
}

func NewHTTPProxy(remoteSocks, cryptoMethod string, password []byte) *HTTPProxy {
	return &HTTPProxy{
		ReverseProxy: &httputil.ReverseProxy{
			Director: director,
			Transport: &http.Transport{
				Dial: func(network, addr string) (net.Conn, error) {
					return dial(network, addr, remoteSocks, cryptoMethod, password)
				},
			},
		},
		remoteServer:       remoteSocks,
		remoteCryptoMethod: cryptoMethod,
		remotePassword:     password,
	}
}

func dial(network, addr, remoteSocks, cryptoMethod string, password []byte) (net.Conn, error) {
	tcpAddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		return nil, err
	}
	remoteSvr, err := NewRemoteSocks(remoteSocks, cryptoMethod, password)
	if err != nil {
		return nil, err
	}

	// version(1) + cmd(1) + reserved(1) + addrType(1) + domainLength(1) + maxDomainLength(256) + port(2)
	req := []byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	copy(req[4:8], []byte(tcpAddr.IP.To4()))
	binary.BigEndian.PutUint16(req[8:10], uint16(tcpAddr.Port))
	err = remoteSvr.Handshake(req)
	if err != nil {
		remoteSvr.Close()
		return nil, err
	}
	conn := &HTTPProxyConn{
		RemoteSocks: remoteSvr,
	}
	return conn, nil
}

func director(request *http.Request) {
	u, err := url.Parse(request.RequestURI)
	if err != nil {
		return
	}
	request.RequestURI = u.RequestURI()
	v := request.Header.Get("Proxy-Connection")
	if v != "" {
		request.Header.Del("Proxy-Connection")
		request.Header.Del("Connection")
		request.Header.Add("Connection", v)
	}
}

func (h *HTTPProxy) Run(addr string) error {
	listen, err := net.Listen("tcp4", addr)
	if err != nil {
		return err
	}
	defer listen.Close()

	return http.Serve(listen, h)
}

func (h *HTTPProxy) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if request.Method == "CONNECT" {
		ServeHTTPTunnel(response, request, h.remoteServer, h.remoteCryptoMethod, h.remotePassword)
	} else {
		h.ReverseProxy.ServeHTTP(response, request)
	}
}

package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type HTTPProxy struct {
	*httputil.ReverseProxy
	connectUpstream ConnectUpstream
}

func NewHTTPProxy(connectUpstream ConnectUpstream) *HTTPProxy {
	return &HTTPProxy{
		ReverseProxy: &httputil.ReverseProxy{
			Director: director,
			Transport: &http.Transport{
				Dial: func(network, addr string) (net.Conn, error) {
					return connectUpstream(addr)
				},
			},
		},
		connectUpstream: connectUpstream,
	}
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

func (h *HTTPProxy) ServeHTTPTunnel(response http.ResponseWriter, request *http.Request) {
	var conn net.Conn
	if hj, ok := response.(http.Hijacker); ok {
		var err error
		if conn, _, err = hj.Hijack(); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(response, "Hijacker failed", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	dest, err := h.connectUpstream(request.Host)
	if err != nil {
		fmt.Fprintf(conn, "HTTP/1.0 500 NewRemoteSocks failed, err:%s\r\n\r\n", err)
		return
	}
	defer dest.Close()

	if request.Body != nil {
		if _, err = io.Copy(dest, request.Body); err != nil {
			fmt.Fprintf(conn, "%d %s", http.StatusBadGateway, err.Error())
			return
		}
	}
	fmt.Fprintf(conn, "HTTP/1.0 200 Connection established\r\n\r\n")

	go io.Copy(dest, conn)
	io.Copy(conn, dest)
}

func (h *HTTPProxy) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if request.Method == "CONNECT" {
		h.ServeHTTPTunnel(response, request)
	} else {
		h.ReverseProxy.ServeHTTP(response, request)
	}
}

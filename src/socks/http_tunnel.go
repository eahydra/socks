package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
)

type HTTPTunnel struct {
	remoteCryptoMethod string
	remotePassword     []byte
	remoteServer       string
}

func NewHTTPTunnel(remoteServer string, remoteCryptoMethod string, remotePassword []byte) *HTTPTunnel {
	return &HTTPTunnel{
		remoteCryptoMethod: remoteCryptoMethod,
		remotePassword:     remotePassword,
		remoteServer:       remoteServer,
	}
}

func (h *HTTPTunnel) Run(addr string) error {
	listener, err := net.Listen("tcp4", addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	return http.Serve(listener, h)
}

func ServeHTTPTunnel(response http.ResponseWriter, request *http.Request,
	remoteServer, remoteCryptoMethod string, remotePassword []byte) {
	if request.Method != "CONNECT" {
		http.Error(response, http.ErrNotSupported.Error(), http.StatusMethodNotAllowed)
		return
	}

	s := strings.Split(request.Host, ":")
	if len(s) == 1 {
		s = append(s, "80")
	}
	if len(s) < 2 {
		ErrLog.Println("Invalid HOST:", request.Host)
		http.Error(response, "Invalid Request", http.StatusBadRequest)
		return
	}
	destHost := s[0]
	destPort, err := strconv.Atoi(s[1])
	if err != nil {
		http.Error(response, "Invalid port", http.StatusBadRequest)
		return
	}

	hj, ok := response.(http.Hijacker)
	if !ok {
		http.Error(response, "Hijacker failed", http.StatusInternalServerError)
		return
	}
	conn, _, err := hj.Hijack()
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	var dest io.ReadWriteCloser
	if remoteServer != "" {
		remoteSvr, err := NewRemoteSocks(remoteServer, remoteCryptoMethod, remotePassword)
		if err != nil {
			ErrLog.Println("HTTPTunnel NewRemoteSocks failed, err:", err)
			fmt.Fprintf(conn, "HTTP/1.0 500 NewRemoteSocks failed, err:%s\r\n\r\n", err)
			return
		}
		defer remoteSvr.Close()

		// version(1) + cmd(1) + reserved(1) + addrType(1) + domainLength(1) + maxDomainLength(256) + port(2)
		req := bytes.NewBuffer(nil)
		req.Write([]byte{0x05, 0x01, 0x00, 0x03})
		req.WriteByte(byte(len(destHost)))
		req.WriteString(destHost)
		buff := []byte{0x00, 0x00}
		binary.BigEndian.PutUint16(buff, uint16(destPort))
		req.Write(buff)

		err = remoteSvr.Handshake(req.Bytes())
		if err != nil {
			ErrLog.Println("HTTPTunnel Handshake with remote server failed, err:", err)
			fmt.Fprintf(conn, "HTTP/1.0 500 Handshake failed, err:%s\r\n\r\n", err)
			return
		}
		dest = remoteSvr

	} else {
		destConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", destHost, destPort))
		if err != nil {
			ErrLog.Println("net.Dial", destHost, destPort, "failed, err:", err)
			fmt.Fprintf(conn, "HTTP/1.0 500 net.Dial failed, err:%s\r\n\r\n", err.Error())
			return
		}
		defer destConn.Close()
		dest = destConn
	}

	if request.Body != nil {
		_, err = io.Copy(dest, request.Body)
		if err != nil {
			fmt.Fprintf(conn, "%d %s", http.StatusBadGateway, err.Error())
			return
		}
	}
	fmt.Fprintf(conn, "HTTP/1.0 200 Connection established\r\n\r\n")

	go io.Copy(dest, conn)
	io.Copy(conn, dest)

}

func (h *HTTPTunnel) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	ServeHTTPTunnel(response, request, h.remoteServer, h.remoteCryptoMethod, h.remotePassword)
}

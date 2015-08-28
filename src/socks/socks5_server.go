package main

import (
	"errors"
	"net"
)

var (
	ErrUnsupportedVersion = errors.New("socks unsupported version")
	ErrUnsupportedCommand = errors.New("socks unsupported command")
	ErrInvalidProtocol    = errors.New("socks invalid protocol")
)

type SOCKS5Server struct {
	cryptoMethod    string
	password        string
	connectUpstream ConnectUpstream
}

func NewSocks5Server(cryptMethod string, password string, connectUpstream ConnectUpstream) *SOCKS5Server {
	return &SOCKS5Server{
		cryptoMethod:    cryptMethod,
		password:        password,
		connectUpstream: connectUpstream,
	}
}

func (s *SOCKS5Server) Run(addr string) error {
	listener, err := net.Listen("tcp4", addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				continue
			} else {
				return err
			}
		}

		if clientConn, err := NewSOCKS5Client(conn, s.cryptoMethod, s.password); err == nil {
			go clientConn.serve(s.connectUpstream)

		} else {
			conn.Close()
		}
	}
	panic("unreached")
}

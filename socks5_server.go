package socks

import (
	"errors"
	"net"
)

var (
	ErrUnsupportedVersion = errors.New("socks unsupported version")
	ErrInvalidProtocol    = errors.New("socks invalid protocol")
)

type SOCKS5Server struct {
	router Router
}

func NewSocks5Server(router Router) *SOCKS5Server {
	return &SOCKS5Server{
		router: router,
	}
}

func (s *SOCKS5Server) Run(listener net.Listener) error {
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

		client := NewSOCKS5Client(conn)
		go client.serve(s.router)
	}
	panic("unreached")
}

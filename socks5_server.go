package socks

import (
	"errors"
	"net"
)

var (
	// ErrUnsupportedVersion means protocol version invalid.
	ErrUnsupportedVersion = errors.New("socks unsupported version")
	// ErrInvalidProtocol means protocol invalid.
	ErrInvalidProtocol = errors.New("socks invalid protocol")
)

// SOCKS5Server implements SOCKS5 Server Protocol, but not support UDP and BIND command.
type SOCKS5Server struct {
	router Router
}

// NewSocks5Server constructs one SOCKS5Server with router.
func NewSocks5Server(router Router) *SOCKS5Server {
	return &SOCKS5Server{
		router: router,
	}
}

// Run begin accept incoming client conn and serve it.
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
}

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
type Socks5Server struct {
	forward Dialer
}

// NewSocks5Server constructs one SOCKS5Server with router.
func NewSocks5Server(forward Dialer) (*Socks5Server, error) {
	return &Socks5Server{
		forward: forward,
	}, nil
}

// Run begin accept incoming client conn and serve it.
func (s *Socks5Server) Serve(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				continue
			} else {
				return err
			}
		}

		go serveSOCKS5Client(conn, s.forward)
	}
}

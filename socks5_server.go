package socks

import "net"

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

		go serveSocks5Client(conn, s.forward)
	}
}

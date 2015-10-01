package socks

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

// SOCKS4Server implements SOCKS4 Server Protocol. but not support udp protocol.
type Socks4Server struct {
	forward Dialer
}

// NewSOCKS4Server constructs one SOCKS4Server
func NewSocks4Server(forward Dialer) (*Socks4Server, error) {
	return &Socks4Server{
		forward: forward,
	}, nil
}

// Run just listen at specify address and serve with incoming new client conn.
func (s *Socks4Server) Serve(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				continue
			} else {
				return err
			}
		}

		go serveSOCKS4Client(conn, s.forward)
	}
}

// SOCKS4Client implement SOCKS4 Client Protocol. It combine with net.Conn,
// so you can use SOCKS4Client as net.Conn to read or write.
type Socks4Client struct {}

func NewSocks4Client(network, address string, forward Dialer) (*Socks4Client, error) {
	// TODO(joseph): Implement it.
	return &Socks4Client{}, nil
}

func (s *Socks4Client) Dial(network, address string) (net.Conn, error) {
	// TODO(joseph): Implement it.
	return nil, nil
}

func serveSOCKS4Client(conn net.Conn, forward Dialer) {
	defer conn.Close()

	cmd, destIP, destPort, err := socks4Handshake(conn)
	if err != nil {
		return
	}

	reply := []byte{0x00, 0x5a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if cmd != 0x01 {
		reply[1] = 0x5b // reject.
		conn.Write(reply)
		return
	}

	dest, err := forward.Dial("tcp", net.JoinHostPort(destIP.String(), fmt.Sprintf("%d", destPort)))
	if err != nil {
		reply[1] = 0x5c // connect failed
		conn.Write(reply)
		return
	}
	defer dest.Close()

	if _, err = conn.Write(reply); err != nil {
		return
	}

	go func() {
		defer conn.Close()
		defer dest.Close()
		io.Copy(dest, conn)
	}()
	io.Copy(conn, dest)
}

func socks4Handshake(conn net.Conn) (cmd byte, ip net.IP, port uint16, err error) {
	// version(1) + command(1) + port(2) + ip(4) + null(1)
	p := [1024]byte{}
	buff := p[:]
	n := 0
	if n, err = io.ReadAtLeast(conn, buff, 8); err != nil {
		return
	}
	if buff[0] != 4 {
		err = ErrUnsupportedVersion
		return
	}
	cmd = buff[1]
	port = binary.BigEndian.Uint16(buff[2:4])
	ip = net.IP(buff[4:8])

	if buff[n-1] != 0 {
		for {
			if n, err = conn.Read(buff); err != nil {
				return
			}
			if buff[n-1] == 0 {
				break
			}
			if err == io.EOF {
				break
			}
		}
	}

	return
}

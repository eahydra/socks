package socks

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

type SOCKS4Server struct {
	router Router
}

func NewSOCKS4Server(router Router) *SOCKS4Server {
	return &SOCKS4Server{
		router: router,
	}
}

func (s *SOCKS4Server) Run(addr string) error {
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
		clientConn := NewSOCKS4ClientConn(conn)
		go clientConn.serve(s.router)
	}
	panic("unreached")
}

type SOCKS4ClientConn struct {
	net.Conn
}

func NewSOCKS4ClientConn(conn net.Conn) *SOCKS4ClientConn {
	clientConn := &SOCKS4ClientConn{
		Conn: conn,
	}
	return clientConn
}

func (c *SOCKS4ClientConn) serve(router Router) {
	defer c.Close()

	cmd, destIP, destPort, err := c.handshake()
	if err != nil {
		return
	}

	reply := []byte{0x00, 0x5a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if cmd != 0x01 {
		reply[1] = 0x5b // reject.
		c.Write(reply)
		return
	}

	dest, err := router.Do(net.JoinHostPort(destIP.String(), fmt.Sprintf("%d", destPort)))
	if err != nil {
		reply[1] = 0x5c // connect failed
		c.Write(reply)
		return
	}
	defer dest.Close()

	if _, err = c.Write(reply); err != nil {
		return
	}

	go func() {
		defer c.Close()
		defer dest.Close()
		io.Copy(dest, c)
	}()
	io.Copy(c, dest)
}

func (c *SOCKS4ClientConn) handshake() (cmd byte, ip net.IP, port uint16, err error) {
	// version(1) + command(1) + port(2) + ip(4) + null(1)
	p := [1024]byte{}
	buff := p[:]
	n := 0
	if n, err = io.ReadAtLeast(c, buff, 8); err != nil {
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
			if n, err = c.Read(buff); err != nil {
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

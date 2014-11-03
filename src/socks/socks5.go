package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

var (
	ErrUnsupportedVersion = errors.New("socks unsupported version")
	ErrUnsupportedCommand = errors.New("socks unsupported command")
	ErrInvalidProtocol    = errors.New("socks invalid protocol")
)

type SOCKS5Server struct {
	localCryptoMethod  string
	localPassword      []byte
	remoteServer       string
	remoteCryptoMethod string
	remotePassowrd     []byte
}

func NewSocks5Server(cryptMethod string, password []byte,
	remoteServer string, remoteCryptoMethod string, remotePassowrd []byte) *SOCKS5Server {
	return &SOCKS5Server{
		localCryptoMethod:  cryptMethod,
		localPassword:      password,
		remoteServer:       remoteServer,
		remoteCryptoMethod: remoteCryptoMethod,
		remotePassowrd:     remotePassowrd,
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
				WarnLog.Println("SOCKS5 listener.Accept temporary error")
				continue
			} else {
				ErrLog.Println("SOCKS5 listener.Accept failed, err:", err.Error())
				return err
			}
		}
		InfoLog.Println("SOCKS5 Incoming new connection, remote:", conn.RemoteAddr().String())

		if clientConn, err := NewSOCKS5ClientConn(conn, s.localCryptoMethod, s.localPassword,
			s.remoteServer, s.remoteCryptoMethod, s.remotePassowrd); err == nil {
			go clientConn.Run()

		} else {
			ErrLog.Println("SOCKS5 NewSOCKS5ClientConn failed, err:", err)
			conn.Close()
		}
	}
	panic("unreached")
}

type SOCKS5ClientConn struct {
	conn net.Conn
	*CipherStream
	remoteServer       string
	remoteCryptoMethod string
	remotePassword     []byte
}

func NewSOCKS5ClientConn(conn net.Conn, localCryptoMethod string, localPassword []byte,
	remoteServer string, remoteCryptoMethod string, remotePassword []byte) (*SOCKS5ClientConn, error) {
	clientConn := &SOCKS5ClientConn{
		conn:               conn,
		remoteServer:       remoteServer,
		remoteCryptoMethod: remoteCryptoMethod,
		remotePassword:     remotePassword,
	}
	var err error
	clientConn.CipherStream, err = NewCipherStream(conn, localCryptoMethod, localPassword)
	if err != nil {
		return nil, err
	}
	return clientConn, nil
}

func (c *SOCKS5ClientConn) Run() {
	defer func() {
		InfoLog.Println("SOCKS5 Connection closed, remote:", c.conn.RemoteAddr().String())
		c.Close()
	}()

	if err := c.handshake(); err != nil {
		WarnLog.Println("SOCKS5 handshake failed, err:", err)
		return
	}

	cmd, destHost, destPort, req, err := c.getCommand()
	if err != nil {
		WarnLog.Println("SOCKS5 getCommand failed, err:", err)
		return
	}
	InfoLog.Printf("SOCKS5 cmd:%d, destHost:%s:%d", cmd, destHost, destPort)
	reply := []byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x22, 0x22}
	if cmd != 0x01 {
		WarnLog.Println("SOCKS5 unsupported command, cmd:", cmd)
		reply[1] = 0x07 // unsupported command
		c.Write(reply)
		return
	}

	var dest io.ReadWriteCloser
	if c.remoteServer != "" {
		remoteSvr, err := NewRemoteSocks(c.remoteServer, c.remoteCryptoMethod, c.remotePassword)
		if err != nil {
			ErrLog.Println("SOCKS5 NewRemoteSocks failed, err:", err)
			reply[1] = 0x05
			c.Write(reply)
			return
		}

		err = remoteSvr.Handshake(req)
		if err != nil {
			ErrLog.Println("SOCKS5 Handshake with remote server failed, err:", err)
			reply[1] = 0x05
			c.Write(reply)
			remoteSvr.Close()
			return
		}
		dest = remoteSvr

	} else {
		destConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", destHost, destPort))
		if err != nil {
			ErrLog.Println("net.Dial", destHost, destPort, "failed, err:", err)
			reply[1] = 0x05
			c.Write(reply)
			destConn.Close()
			return
		}
		dest = destConn
	}
	defer dest.Close()

	reply[1] = 0x00
	if _, err = c.Write(reply); err != nil {
		ErrLog.Println("SOCKS5 write succeed reply failed. err:", err)
		return
	}

	go func() {
		defer c.Close()
		defer dest.Close()

		io.Copy(c, dest)
	}()

	io.Copy(dest, c)
}

func (c *SOCKS5ClientConn) handshake() error {
	// version(1) + numMethods(1) + [256]methods
	buff := make([]byte, 258)
	n, err := io.ReadAtLeast(c, buff, 2)
	if err != nil {
		return err
	}
	if buff[0] != 5 {
		return ErrUnsupportedVersion
	}
	numMethod := int(buff[1])
	numMethod += 2
	if n < numMethod {
		if _, err = io.ReadFull(c, buff[n:numMethod]); err != nil {
			return err
		}
	} else if n > numMethod {
		return ErrInvalidProtocol
	}

	buff[1] = 0 // no authentication
	if _, err := c.Write(buff[:2]); err != nil {
		return err
	}
	return nil
}

func (c *SOCKS5ClientConn) getCommand() (cmd byte, destHost string, destPort uint16, data []byte, err error) {
	// version(1) + cmd(1) + reserved(1) + addrType(1) + domainLength(1) + maxDomainLength(256) + port(2)
	buff := make([]byte, 263)
	var n int
	n, err = io.ReadAtLeast(c, buff, 9)
	if err != nil {
		return
	}
	if buff[0] != 5 {
		err = ErrUnsupportedVersion
		return
	}
	cmd = buff[1]
	totalLength := 0
	switch buff[3] {
	case 0x01: // IPV4
		totalLength = 3 + 1 + 4 + 2 // version + cmd + reserved + addrType + ip + 2
	case 0x03: // Domain
		totalLength = 3 + 1 + 1 + int(buff[4]) + 2 // ver + cmd + reserved + addrType + domainLength + Length + 2
	case 0x04: // IPV6
		totalLength = 3 + 1 + 16 + 2 // version + cmd + reserved + addrType + ipv6 + 2
	}
	if n < totalLength {
		if _, err = io.ReadFull(c, buff[n:totalLength]); err != nil {
			return
		}
	} else if n > totalLength {
		err = ErrInvalidProtocol
		return
	}
	switch buff[3] {
	case 0x01:
		destHost = net.IP(buff[4 : 4+net.IPv4len]).String()
	case 0x03:
		destHost = string(buff[5 : 5+int(buff[4])])
	case 0x04:
		destHost = net.IP(buff[4 : 4+net.IPv6len]).String()
	}
	destPort = binary.BigEndian.Uint16(buff[totalLength-2 : totalLength])
	data = buff[:totalLength]
	return
}

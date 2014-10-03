package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

type SOCKS4Server struct {
	remoteServer       string
	remoteCryptoMethod string
	remotePassword     []byte
}

func NewSOCKS4Server(remoteServer string, cryptMethod string, password []byte) *SOCKS4Server {
	return &SOCKS4Server{
		remoteServer:       remoteServer,
		remoteCryptoMethod: cryptMethod,
		remotePassword:     password,
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
				WarnLog.Println("SOCKS4 listener.Accept temporary error")
				continue
			} else {
				ErrLog.Println("SOCKS4 listener.Accept failed, err:", err.Error())
				return err
			}
		}
		InfoLog.Println("SOCKS4 Incoming new connection, remote:", conn.RemoteAddr().String())
		clientConn := NewSOCKS4ClientConn(conn, s.remoteServer, s.remoteCryptoMethod, s.remotePassword)
		go clientConn.Run()
	}
	panic("unreached")
}

type SOCKS4ClientConn struct {
	net.Conn
	remoteServer string
	cryptoMethod string
	password     []byte
}

func NewSOCKS4ClientConn(conn net.Conn, remoteServer, cryptoMethod string, password []byte) *SOCKS4ClientConn {
	clientConn := &SOCKS4ClientConn{
		Conn:         conn,
		remoteServer: remoteServer,
		cryptoMethod: cryptoMethod,
		password:     password,
	}
	return clientConn
}

func (c *SOCKS4ClientConn) Run() {
	defer func() {
		InfoLog.Println("SOCKS4 Connection closed, remote:", c.RemoteAddr().String())
		c.Close()
	}()

	cmd, destPort, destHost, err := c.handshake()
	if err != nil {
		WarnLog.Println("SOCKS4 handshake failed, err:", err)
		return
	}
	InfoLog.Printf("SOCKS4 cmd:%d, destHost:%s:%d", cmd, destHost, destPort)

	reply := []byte{0x00, 0x5a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if cmd != 0x01 {
		WarnLog.Println("SOCKS4 unsupported command, cmd:", cmd)
		reply[1] = 0x5b // reject
		c.Write(reply)
		return
	}

	var dest io.ReadWriteCloser
	if c.remoteServer != "" {
		remoteSvr, err := NewRemoteSocks(c.remoteServer, c.cryptoMethod, c.password)
		if err != nil {
			ErrLog.Println("SOCKS4 NewRemoteSocks failed, err:", err)
			reply[1] = 0x5c
			c.Write(reply)
			return
		}
		defer remoteSvr.Close()

		// version(1) + cmd(1) + reserved(1) + addrType(1) + domainLength(1) + maxDomainLength(256) + port(2)
		req := []byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		copy(req[4:8], []byte(net.ParseIP(destHost).To4()))
		binary.BigEndian.PutUint16(req[8:10], destPort)

		err = remoteSvr.Handshake(req)
		if err != nil {
			ErrLog.Println("SOCKS4 Handshake with remote server failed, err:", err)
			reply[1] = 0x5c
			c.Write(reply)
			return
		}
		dest = remoteSvr

	} else {
		destConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", destHost, destPort))
		if err != nil {
			ErrLog.Println("net.Dial", destHost, destPort, "failed, err:", err)
			reply[1] = 0x5c
			c.Write(reply)
			return
		}
		defer destConn.Close()
		dest = destConn
	}

	if _, err = c.Write(reply); err != nil {
		ErrLog.Println("SOCKS4 write succeed reply failed. err:", err)
		return
	}

	go io.Copy(dest, c)
	_, err = io.Copy(c, dest)
}

func (c *SOCKS4ClientConn) handshake() (cmd byte, port uint16, ip string, err error) {
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
	ip = net.IP(buff[4:8]).String()

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

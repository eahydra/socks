package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

var (
	ErrUnsupportedVersion = errors.New("socks unsupported version")
	ErrUnsupportedCommand = errors.New("socks unsupported command")
	ErrInvalidProtocol    = errors.New("socks invalid protocol")
)

var (
	InfoLog = log.New(os.Stdout, "INFO  ", log.LstdFlags)
	ErrLog  = log.New(os.Stderr, "ERROR ", log.LstdFlags)
	WarnLog = log.New(os.Stdout, "WARN  ", log.LstdFlags)
)

type ClientConn struct {
	conn net.Conn
	*CipherStream
	cryptoMethod string
	password     []byte
}

func NewClientConn(conn net.Conn, localNeedCrypto bool, cryptoMethod string, password []byte) (*ClientConn, error) {
	clientConn := &ClientConn{
		conn:         conn,
		cryptoMethod: cryptoMethod,
		password:     password,
	}
	if !localNeedCrypto {
		cryptoMethod = ""
	}
	var err error
	clientConn.CipherStream, err = NewCipherStream(conn, cryptoMethod, password)
	if err != nil {
		return nil, err
	}
	return clientConn, nil
}

func (c *ClientConn) Run(remoteSocks string) {
	defer func() {
		InfoLog.Println("Connection closed, remote:", c.conn.RemoteAddr().String())
		c.Close()
	}()

	if err := handshake(c); err != nil {
		WarnLog.Println("handshake failed, err:", err)
		return
	}

	cmd, destHost, destPort, req, err := getCommand(c)
	if err != nil {
		WarnLog.Println("getCommand failed, err:", err)
		return
	}
	reply := []byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x22, 0x22}
	if cmd != 0x01 {
		WarnLog.Println("unsupported command, cmd:", cmd)
		reply[1] = 0x07 // unsupported command
		c.Write(reply)
		return
	}

	var dest io.ReadWriteCloser
	if remoteSocks != "" {
		remoteSvr, err := NewRemoteSocks(remoteSocks, c.cryptoMethod, c.password)
		if err != nil {
			ErrLog.Println("NewRemoteSocks failed, err:", err)
			reply[1] = 0x05
			c.Write(reply)
			return
		}
		defer remoteSvr.Close()

		err = remoteSvr.Handshake(req)
		if err != nil {
			ErrLog.Println("Handshake with remote server failed, err:", err)
			reply[1] = 0x05
			c.Write(reply)
			return
		}
		dest = remoteSvr

	} else {
		destConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", destHost, destPort))
		if err != nil {
			ErrLog.Println("net.Dial", destHost, destPort, "failed, err:", err)
			reply[1] = 0x05
			c.Write(reply)
			return
		}
		defer destConn.Close()
		dest = destConn
	}

	reply[1] = 0x00
	if _, err = c.Write(reply); err != nil {
		ErrLog.Println("write succeed reply failed. err:", err)
		return
	}

	if dest == nil {
		panic("dest is nil")
	}

	go io.Copy(dest, c)
	io.Copy(c, dest)
}

func handshake(rw io.ReadWriter) error {
	// version(1) + numMethods(1) + [256]methods
	buff := make([]byte, 258)
	n, err := io.ReadAtLeast(rw, buff, 2)
	if err != nil {
		return err
	}
	if buff[0] != 5 {
		return ErrUnsupportedVersion
	}
	numMethod := int(buff[1])
	numMethod += 2
	if n < numMethod {
		if _, err = io.ReadFull(rw, buff[n:numMethod]); err != nil {
			return err
		}
	} else if n > numMethod {
		return ErrInvalidProtocol
	}

	buff[1] = 0 // no authentication
	if _, err := rw.Write(buff[:2]); err != nil {
		return err
	}
	return nil
}

func getCommand(reader io.Reader) (cmd byte, destHost string, destPort uint16, data []byte, err error) {
	// version(1) + cmd(1) + reserved(1) + addrType(1) + domainLength(1) + maxDomainLength(256) + port(2)
	buff := make([]byte, 263)
	var n int
	n, err = io.ReadAtLeast(reader, buff, 9)
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
		if _, err = io.ReadFull(reader, buff[n:totalLength]); err != nil {
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

func main() {
	globalCfg, err := LoadConfig()
	if err != nil {
		ErrLog.Println("LoadConfig failed, err:", err)
		return
	}
	InfoLog.Println(globalCfg)

	listener, err := net.Listen("tcp4", fmt.Sprintf("%s:%d", globalCfg.ListenIP, globalCfg.ListenPort))
	if err != nil {
		ErrLog.Println("net.Listen failed. err:", err)
		return
	}
	defer listener.Close()

	password := []byte(globalCfg.CryptoPassword)
	remoteSocks := ""
	if globalCfg.RemoteSocksIP != "" && globalCfg.RemoteSocksPort != 0 {
		remoteSocks = fmt.Sprintf("%s:%d", globalCfg.RemoteSocksIP, globalCfg.RemoteSocksPort)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				WarnLog.Println("listener.Accept temporary error")
				continue
			} else {
				ErrLog.Println("listener.Accept failed, err:", err.Error())
				return
			}
		}
		InfoLog.Println("Incoming new connection, remote:", conn.RemoteAddr().String())
		if clientConn, err := NewClientConn(conn, globalCfg.LocalNeedCrypto, globalCfg.CryptoMethod, password); err == nil {
			go clientConn.Run(remoteSocks)

		} else {
			ErrLog.Println("NewClientConn failed, err:", err)
			conn.Close()
		}
	}
}

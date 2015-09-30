package socks

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
)

// SOCKS5Client implements SOCKS5 Client Protocol.
// You can use it as net.Conn to read/write.
type SOCKS5Client struct {
	net.Conn
}

// NewSOCKS5Client constructs one SOCKS5Client
// Call this function with conn that accept from net.Listener or from net.Dial
func NewSOCKS5Client(conn net.Conn) *SOCKS5Client {
	return &SOCKS5Client{
		Conn: conn,
	}
}

func parseAddress(address string) (interface{}, string, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, "", err
	}
	ip := net.ParseIP(address)
	if ip != nil {
		return ip, port, nil
	}
	return host, port, nil
}

func buildSOCKS5Request(addr string) ([]byte, error) {
	host, p, err := parseAddress(addr)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return nil, err
	}

	// version(1) + cmd(1) + reserved(1) + addrType(1) + domainLength(1) + maxDomainLength(256) + port(2)
	req := bytes.NewBuffer(nil)
	req.Write([]byte{0x05, 0x01, 0x00})
	switch h := host.(type) {
	case string:
		{
			req.WriteByte(0x03)
			req.WriteByte(byte(len(h)))
			req.WriteString(h)
		}
	case net.IP:
		{
			if len(h) == net.IPv4len {
				req.WriteByte(0x01)
			} else {
				req.WriteByte(0x04)
			}
			req.Write([]byte(h))
		}
	}
	binary.Write(req, binary.BigEndian, uint16(port))
	return req.Bytes(), nil
}

// RequestProxy send SOCKS5 Proxy request with address to remote peer.
func (c *SOCKS5Client) RequestProxy(address string) error {
	buff := []byte{0x05, 0x01, 0x00}
	_, err := c.Write(buff)
	if err != nil {
		return err
	}
	n, err := c.Read(buff[:2])
	buff = buff[:n]
	if err != nil {
		return err
	}
	if buff[0] != 0x05 || buff[1] != 0x00 {
		return ErrInvalidProtocol
	}
	cmd, err := buildSOCKS5Request(address)
	if err != nil {
		return err
	}
	_, err = c.Write(cmd)
	if err != nil {
		return err
	}
	var reply [10]byte
	_, err = c.Read(reply[:])
	if err != nil {
		return err
	}
	return nil
}

func (c *SOCKS5Client) serve(router Router) {
	defer c.Close()

	if err := c.handshake(); err != nil {
		return
	}

	cmd, destHost, destPort, err := c.getCommand()
	if err != nil {
		return
	}
	reply := []byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x22, 0x22}
	if cmd != 0x01 {
		reply[1] = 0x07 // unsupported command
		c.Write(reply)
		return
	}

	dest, err := router.Do(net.JoinHostPort(destHost, fmt.Sprintf("%d", destPort)))
	if err != nil {
		reply[1] = 0x05
		c.Write(reply)
		return
	}
	defer dest.Close()

	reply[1] = 0x00
	if _, err = c.Write(reply); err != nil {
		return
	}

	go func() {
		defer c.Close()
		defer dest.Close()
		io.Copy(c, dest)
	}()

	io.Copy(dest, c)
}

func (c *SOCKS5Client) handshake() error {
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

func (c *SOCKS5Client) getCommand() (cmd byte, destHost string, destPort uint16, err error) {
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
	return
}

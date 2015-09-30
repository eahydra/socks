package socks

import (
	"bytes"
	"encoding/binary"
	"net"
	"strconv"
)

// ShadowSocksClient implements ShadowSocks Client Protocol and combine with net.Conn
// so you can use ShadowSocksClient as net.Conn to read or write.
type ShadowSocksClient struct {
	net.Conn
}

// NewShadowSocksClient constructs one ShadowSocksClient.
// Call this function with conn that accept from net.Listener or from net.Dial
func NewShadowSocksClient(conn net.Conn) *ShadowSocksClient {
	return &ShadowSocksClient{
		Conn: conn,
	}
}

// RequestProxy send ShadowSocks proxy request with addr to remote peer.
func (s *ShadowSocksClient) RequestProxy(addr string) error {
	req, err := buildShadowSocksRequest(addr)
	if err != nil {
		return err
	}
	_, err = s.Write(req)
	if err != nil {
		return err
	}
	return nil
}

func buildShadowSocksRequest(address string) ([]byte, error) {
	host, p, err := parseAddress(address)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return nil, err
	}
	req := bytes.NewBuffer(nil)
	switch host.(type) {
	case string:
		domain := host.(string)
		req.WriteByte(3)
		req.WriteByte(byte(len(domain)))
		req.WriteString(domain)
		binary.Write(req, binary.BigEndian, uint16(port))

	case net.IP:
		ip := host.(net.IP)
		if len(ip) == net.IPv4len {
			req.WriteByte(1)
		} else {
			req.WriteByte(4)
		}
		req.WriteByte(byte(len(ip)))
		req.Write(ip)
		binary.Write(req, binary.BigEndian, uint16(port))
	}
	return req.Bytes(), nil
}

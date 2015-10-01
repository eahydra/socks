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
	network string
	address string
	forward Dialer
}

// NewShadowSocksClient constructs one ShadowSocksClient.
// Call this function with conn that accept from net.Listener or from net.Dial
func NewShadowSocksClient(network, address string, forward Dialer) (*ShadowSocksClient, error) {
	return &ShadowSocksClient{
		network: network,
		address: address,
		forward: forward,
	}, nil
}

func (s *ShadowSocksClient) Dial(network, address string) (net.Conn, error) {
	conn, err := s.forward.Dial(s.network, s.address)
	if err != nil {
		return nil, err
	}
	connClose := &conn
	defer func() {
		if connClose != nil {
			(*connClose).Close()
		}
	}()

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
	_, err = conn.Write(req.Bytes())
	if err != nil {
		return nil, err
	}
	connClose = nil
	return conn, nil
}

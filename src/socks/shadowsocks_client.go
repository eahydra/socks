package main

import (
	"bytes"
	"encoding/binary"
	"net"
	"time"
)

type ShadowSocksClient struct {
	conn net.Conn
	CipherStreamReadWriter
}

func DialShadowSocks(addr, cryptoMethod, password string) (*ShadowSocksClient, error) {
	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		return nil, err
	}
	cipher, err := NewCipherStream(conn, cryptoMethod, []byte(password))
	if err != nil {
		conn.Close()
		return nil, err
	}
	return &ShadowSocksClient{
		conn: conn,
		CipherStreamReadWriter: cipher,
	}, nil
}

func buildShadowSocksRequest(addr string) ([]byte, error) {
	host, port, err := parseAddress(addr)
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

func (s *ShadowSocksClient) ConnectUpstream(addr string) error {
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

func (s *ShadowSocksClient) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

func (s *ShadowSocksClient) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

func (s *ShadowSocksClient) SetDeadline(t time.Time) error {
	return s.conn.SetDeadline(t)
}

func (s *ShadowSocksClient) SetReadDeadline(t time.Time) error {
	return s.conn.SetReadDeadline(t)
}

func (s *ShadowSocksClient) SetWriteDeadline(t time.Time) error {
	return s.conn.SetWriteDeadline(t)
}

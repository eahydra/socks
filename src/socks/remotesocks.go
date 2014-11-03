package main

import (
	"net"
)

type RemoteSocks struct {
	conn net.Conn
	*CipherStream
}

func NewRemoteSocks(remote string, cryptoMethod string, password []byte) (*RemoteSocks, error) {
	conn, err := net.Dial("tcp", remote)
	if err != nil {
		return nil, err
	}

	stream, err := NewCipherStream(conn, cryptoMethod, password)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return &RemoteSocks{
		conn:         conn,
		CipherStream: stream,
	}, nil
}

func (r *RemoteSocks) Handshake(cmd []byte) error {
	buff := []byte{0x05, 0x01, 0x00}
	_, err := r.Write(buff)
	if err != nil {
		return err
	}
	n, err := r.Read(buff[:2])
	buff = buff[:n]
	if err != nil {
		return err
	}
	if buff[0] != 0x05 || buff[1] != 0x00 {
		return ErrInvalidProtocol
	}
	_, err = r.Write(cmd)
	if err != nil {
		return err
	}
	var reply [10]byte
	_, err = r.Read(reply[:])
	if err != nil {
		return err
	}
	return nil
}

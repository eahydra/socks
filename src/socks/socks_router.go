package main

import "net"

type SOCKClient interface {
	net.Conn
	RequestProxy(address string) error
}

type ClientFactory func(conn net.Conn) SOCKClient

type SOCKSRouter struct {
	serverAddress string
	clientFactory ClientFactory
	decorators    []ConnDecorator
}

func NewSOCKSRouter(serverAddress string, factory ClientFactory, ds ...ConnDecorator) *SOCKSRouter {
	s := &SOCKSRouter{
		serverAddress: serverAddress,
		clientFactory: factory,
	}
	s.decorators = append(s.decorators, ds...)
	return s
}

func (s *SOCKSRouter) Do(address string) (net.Conn, error) {
	conn, err := net.Dial("tcp", s.serverAddress)
	if err != nil {
		return nil, err
	}
	dconn, err := DecorateConn(conn, s.decorators...)
	if err != nil {
		dconn.Close()
		return nil, err
	}
	client := s.clientFactory(dconn)
	err = client.RequestProxy(address)
	if err != nil {
		dconn.Close()
		return nil, err
	}
	return client, nil
}

package main

import (
	"net"

	"github.com/eahydra/socks"
)

type SOCKClient interface {
	net.Conn
	RequestProxy(address string) error
}

type ClientFactory func(conn net.Conn) SOCKClient

type SOCKSRouter struct {
	serverAddress string
	baseRouter    socks.Router
	clientFactory ClientFactory
	decorators    []ConnDecorator
}

func NewSOCKSRouter(serverAddress string, baseRouter socks.Router, factory ClientFactory, ds ...ConnDecorator) *SOCKSRouter {
	s := &SOCKSRouter{
		serverAddress: serverAddress,
		baseRouter:    baseRouter,
		clientFactory: factory,
	}
	s.decorators = append(s.decorators, ds...)
	return s
}

func (s *SOCKSRouter) Do(address string) (net.Conn, error) {
	conn, err := s.baseRouter.Do(s.serverAddress)
	if err != nil {
		return nil, err
	}
	dconn, err := DecorateConn(conn, s.decorators...)
	if err != nil {
		conn.Close()
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

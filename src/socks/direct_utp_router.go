package main

import (
	"net"

	"github.com/anacrolix/utp"
)

type DirectUTPRouter struct{}

func NewDirectUTPRouter() *DirectUTPRouter {
	return &DirectUTPRouter{}
}

func (d *DirectUTPRouter) Do(address string) (net.Conn, error) {
	return utp.Dial(address)
}

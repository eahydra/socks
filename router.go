package socks

import "net"

type Router interface {
	Do(address string) (net.Conn, error)
}

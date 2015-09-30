package socks

import "net"

// Router just defines connect operation.
// Pass the server's address, if succeeded, return net.Conn, otherwise return nil and error code.
type Router interface {
	Do(address string) (net.Conn, error)
}

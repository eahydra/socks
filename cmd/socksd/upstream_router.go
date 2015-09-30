package main

import (
	"net"

	"github.com/eahydra/socks"
)

type RouterBalancer interface {
	GetNextRouter() socks.Router
}

type UpstreamRouter struct {
	routerBalancer RouterBalancer
}

func NewUpstreamRouter(balancer RouterBalancer) *UpstreamRouter {
	return &UpstreamRouter{
		routerBalancer: balancer,
	}
}

func (u *UpstreamRouter) Do(address string) (net.Conn, error) {
	router := u.routerBalancer.GetNextRouter()
	return router.Do(address)
}

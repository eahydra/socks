package main

import "net"

type Router interface {
	Do(address string) (net.Conn, error)
}

type RouterBalancer interface {
	GetNextRouter() Router
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

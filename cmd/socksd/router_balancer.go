package main

import (
	"sync"

	"github.com/eahydra/socks"
)

type UpstreamRouterBalancer struct {
	routers    []socks.Router
	nextRouter int
	lock       sync.Mutex
}

func NewUpstreamRouterBalancer(routers []socks.Router) *UpstreamRouterBalancer {
	return &UpstreamRouterBalancer{
		routers: routers,
	}
}

func (u *UpstreamRouterBalancer) GetNextRouter() socks.Router {
	u.lock.Lock()
	defer u.lock.Unlock()
	index := u.nextRouter
	u.nextRouter++
	if u.nextRouter >= len(u.routers) {
		u.nextRouter = 0
	}
	if index < len(u.routers) {
		return u.routers[index]
	}
	panic("unreached")
}

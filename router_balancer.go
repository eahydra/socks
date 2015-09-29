package main

import "sync"

type UpstreamRouterBalancer struct {
	routers    []Router
	nextRouter int
	lock       sync.Mutex
}

func NewUpstreamRouterBalancer(routers []Router) *UpstreamRouterBalancer {
	return &UpstreamRouterBalancer{
		routers: routers,
	}
}

func (u *UpstreamRouterBalancer) GetNextRouter() Router {
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

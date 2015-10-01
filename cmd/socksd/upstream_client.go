package main

import (
	"net"
	"sync"

	"github.com/eahydra/socks"
)

type UpstreamDialer struct {
	forwardDialers []socks.Dialer
	nextRouter     int
	lock           sync.Mutex
}

func NewUpstreamDialer(forwardDialers []socks.Dialer) *UpstreamDialer {
	return &UpstreamDialer{
		forwardDialers: forwardDialers,
	}
}

func (u *UpstreamDialer) getNextDialer() socks.Dialer {
	u.lock.Lock()
	defer u.lock.Unlock()
	index := u.nextRouter
	u.nextRouter++
	if u.nextRouter >= len(u.forwardDialers) {
		u.nextRouter = 0
	}
	if index < len(u.forwardDialers) {
		return u.forwardDialers[index]
	}
	panic("unreached")
}

func (u *UpstreamDialer) Dial(network, address string) (net.Conn, error) {
	router := u.getNextDialer()
	conn, err := router.Dial(network, address)
	if err != nil {
		ErrLog.Println("UpstreamDialer router.Dial failed, err:", err, network, address)
		return nil, err
	}
	return conn, nil
}

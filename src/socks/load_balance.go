package main

import (
	"sync/atomic"
)

type LoadBalancer func() (addr, method, password string)

func NewLoadBalancer(remotes []RemoteConfig) LoadBalancer {
	if len(remotes) == 0 {
		remotes = append(remotes, RemoteConfig{})
	}
	var currentConfig int32
	return func() (addr, method, password string) {
		index := atomic.AddInt32(&currentConfig, 1)
		if index >= int32(len(remotes)) {
			index = 0
			atomic.StoreInt32(&currentConfig, 0)
		}
		return remotes[index].RemoteServer, remotes[index].RemoteCryptoMethod, remotes[index].RemotePassword
	}
}

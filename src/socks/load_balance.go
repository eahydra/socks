package main

import (
	"sync/atomic"
)

type LoadBalancer func() (serverType, addr, method, password string)

func NewLoadBalancer(allUpstreamConfig []UpstreamConfig) LoadBalancer {
	if len(allUpstreamConfig) == 0 {
		allUpstreamConfig = append(allUpstreamConfig, UpstreamConfig{})
	}
	var currentConfig int32
	return func() (serverType, addr, method, password string) {
		index := atomic.AddInt32(&currentConfig, 1)
		if index >= int32(len(allUpstreamConfig)) {
			index = 0
			atomic.StoreInt32(&currentConfig, 0)
		}
		serverType = allUpstreamConfig[index].ServerType
		addr = allUpstreamConfig[index].Addr
		method = allUpstreamConfig[index].CryptoMethod
		password = allUpstreamConfig[index].Password
		return
	}
}

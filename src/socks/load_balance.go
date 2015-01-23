package main

import (
	"sync/atomic"
)

type LoadBalancer func() (addr, method, password string)

func NewLoadBalancer(allUpstreamConfig []UpstreamConfig) LoadBalancer {
	if len(allUpstreamConfig) == 0 {
		allUpstreamConfig = append(allUpstreamConfig, UpstreamConfig{})
	}
	var currentConfig int32
	return func() (addr, method, password string) {
		index := atomic.AddInt32(&currentConfig, 1)
		if index >= int32(len(allUpstreamConfig)) {
			index = 0
			atomic.StoreInt32(&currentConfig, 0)
		}
		return allUpstreamConfig[index].Addr, allUpstreamConfig[index].CryptoMethod, allUpstreamConfig[index].Password
	}
}

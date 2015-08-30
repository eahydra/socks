package main

import (
	"encoding/json"
	"io/ioutil"
)

type UpstreamConfig struct {
	ServerType   string `json:"serverType"`
	CryptoMethod string `json:"cryptoMethod"`
	Password     string `json:"password"`
	Addr         string `json:"addr"`
}

type Config struct {
	HTTPProxyAddr       string           `json:"httpProxyAddr"`
	SOCKS4Addr          string           `json:"socks4Addr"`
	SOCKS5Addr          string           `json:"socks5Addr"`
	UTPSOCKS5Addr       string           `json:"utpsocks5Addr"`
	LocalCryptoMethod   string           `json:"localCryptoMethod"`
	LocalCryptoPassword string           `json:"localPassword"`
	DNSCacheTimeout     int              `json:dnsCacheTimeout`
	AllUpstreamConfig   []UpstreamConfig `json:"upstream"`
}

type ConfigGroup struct {
	PprofAddr string   `json:"pprof"`
	AllConfig []Config `json:"configs"`
}

func (c *Config) String() string {
	data, _ := json.Marshal(c)
	return string(data)
}

func (c *ConfigGroup) String() string {
	data, _ := json.Marshal(c)
	return string(data)
}

func LoadConfigGroup(s string) (*ConfigGroup, error) {
	data, err := ioutil.ReadFile(s)
	if err != nil {
		return nil, err
	}
	cfgGroup := &ConfigGroup{}
	if err = json.Unmarshal(data, cfgGroup); err != nil {
		return nil, err
	}
	return cfgGroup, nil
}

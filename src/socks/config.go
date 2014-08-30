package main

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	ListenIP        string `json:"listenIP"`
	ListenPort      uint32 `json:"listenPort"`
	LocalNeedCrypto bool   `json:"localNeedCrypto"`
	CryptoMethod    string `json:"cryptoMethod"`
	CryptoPassword  string `json:"cryptoPassword"`
	RemoteSocksIP   string `json:"remoteSocksIP"`
	RemoteSocksPort uint32 `json:"remoteSocksPort"`
}

func LoadConfig() (*Config, error) {
	data, err := ioutil.ReadFile("./.config")
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	if err = json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) String() string {
	data, _ := json.Marshal(c)
	return string(data)
}

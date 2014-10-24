package main

func main() {
	conf, err := LoadConfig("socks.config")
	if err != nil {
		ErrLog.Println("initGlobalConfig failed, err:", err)
		return
	}
	InfoLog.Println(conf)

	remoteServer := conf.RemoteServer
	remoteCryptoMethod := conf.RemoteCryptoMethod
	remotePassword := []byte(conf.RemotePassword)

	httpProxy := NewHTTPProxy(remoteServer, remoteCryptoMethod, remotePassword)
	go httpProxy.Run(conf.HTTPProxyAddr)

	socks4Svr := NewSOCKS4Server(remoteServer, remoteCryptoMethod, remotePassword)
	go socks4Svr.Run(conf.SOCKS4Addr)

	socks5Svr := NewSocks5Server(conf.LocalCryptoMethod, []byte(conf.LocalCryptoPassword),
		remoteServer, remoteCryptoMethod, remotePassword)
	socks5Svr.Run(conf.SOCKS5Addr)
}

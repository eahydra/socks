### Description

Socks is a proxy server. Support SOCK4, SOCK5, HTTP Tunnel and HTTP Proxy.
It supports socks5 upstream. You can deploy the same application in another machine.
If you use upstream future, you can set crypto method and password. Now just supports RC4, DES.

### Config  

The config file must named **socks.config**. The file must be deploy with socks file. The format is shown below:
```json

{
	"httpProxyAddr":":36663",
	"socks4Addr":":36665",
	"socks5Addr":":7777",
						
	"localCryptoMethod":"",
	"localPassword":"",

	"remote":{
		"remoteCryptoMethod":"rc4",
		"remotePassword":"123456",
		"remoteServer":"UpstreamServerIP:9999"
	}
}
```

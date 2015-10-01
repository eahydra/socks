# SOCKS
[![Build Status](https://travis-ci.org/eahydra/socks.svg?branch=master)](https://travis-ci.org/eahydra/socks)  [![GoDoc](https://godoc.org/github.com/eahydra/socks?status.svg)](https://godoc.org/github.com/eahydra/socks)

SOCKS implements SOCKS4/5 Proxy Protocol and HTTP Tunnel which can help you get through firewall.
The cmd/socksd build with package SOCKS, supports cipher connection which crypto method is rc4, des, aes-128-cfb, aes-192-cfb and aes-256-cfb, upstream which can be shadowsocks or socsk5.

# Install
Assume you have go installed, you can install from source.
```
go get github.com/eahydra/socks/cmd/socksd
```

# Usage
Configuration file is in json format. The file must name **socks.config** and put it with socksd together.
Configuration parameters as follows:
```json
{
    "pprof": ":7171",
    "configs":[
        {
	        "httpProxyAddr":":36663",
	        "socks4Addr": ":36665",
	        "socks5Addr": ":7777",
	        "localCryptoMethod": "",
	        "localPassword": "",
	        "dnsCacheTimeout":10,
	        "upstream": [
		        {
                    "serverType":"socks5",
			        "cryptoMethod": "rc4",
			        "password": "abcd#1234",
			        "addr": "54.64.248.242:9999"
		        },
		        {
		        	"serverType":"shadowsocks"
			        "cryptoMethod": "aes-256-cfb",
			        "password": "abcd#1234",
			        "addr": "54.64.214.156:9999"
		        }
            ]
        }
    ]
}

```

*  **pprof**               	- Used to start go pprof to debug   
*  **configs**             	- The array of proxy config item    
*  **httpProxyAddr**       	- (OPTIONAL) Enable http proxy tunnel (127.0.0.1:8080 or :8080)   
*  **socks4Addr**          	- (OPTIONAL) Enable SOCKS4 proxy (127.0.0.1:9090 or :9090)   
*  **socks5Addr**          	- (OPTIONAL) Enable SOCKS5 proxy (127.0.0.1:9999 or :9999)   
*  **localCryptoMethod**   	- (OPTIONAL) SOCKS5's crypto method, now supports rc4, des, aes-128-cfb, aes-192-cfb and aes-256-cfb   
*  **localPasssword**      	- If you set **localCryptoMethod**, you must also set passsword   
*  **dnsCacheTimeout**     	- (OPTIONAL) Enable dns cache (unit is second)   
*  **upstream**            	- Specifies the upstream proxy servers   
*  **serverType**         	- Specifies the type of upstream proxy server. Now supports shadowsocks and socks5   
*  **cryptoMethod**        	- Specifies the crypto method of upstream proxy server. The crypto method is same as **localCryptoMethod**   
*  **password**            	- Specifies the crypto password of upstream proxy server  
*  **addr**                	- Specifies the address of upstream proxy server (8.8.8.8:1111)  
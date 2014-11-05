### Description

Socks is a proxy server. Support SOCK4, SOCK5, HTTP Tunnel and HTTP Proxy.
It supports socks5 upstream. You can deploy the same application in another machines. And support load balance.
If you use upstream future, you can set crypto method and password. Now just supports RC4, DES.

### Config  

The config file must named **socks.config**. The file must be deploy with socks file. The format is shown below:
```json

{
    "pprof": ":7171",
    "httpProxyAddr": ":36663",
    "socks4Addr": ":36665",
    "socks5Addr": ":7777",
    "localCryptoMethod": "",
    "localPassword": "",
    "remotes": [
        {
            "remoteCryptoMethod": "rc4",
            "remotePassword": "abcd#1234",
            "remoteServer": "11.22.33.44:9999"
        },
        {
            "remoteCryptoMethod": "rc4",
            "remotePassword": "abcd#1234",
            "remoteServer": "55.66.77.88:9999"
        }
    ]
}
```

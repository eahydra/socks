### Description

Socks is a proxy server. Support SOCKS4, SOCKS5, HTTP Tunnel and HTTP Proxy.
### 说明
这是一个代理服务器，支持SOCK4,SOCK5以及HTTP PROXY。可以用于科学上网。
既可以部署在本地作为一个简单的端口转发，也可以部署在服务器上做为一个代理服务器。
同时，还支持多个UPSTREAM，目前仅支持SOCK5的UPSTREAM。针对多个UPSTREAM可以做最简单的round-robin负载均衡。
为了确保能够科学上网，针对从本地到UPSTREAM的数据可以做相应的加密，目前支持的加密算法仅仅是RC4和DES，足够一般使用。
如果部署在服务器上作为一个代理服务器，内部还支持DNS缓存，当然你也可以简单的配置对应的超时时间来清除缓存。


### 部署  
git clone https://github.com/eahydra/socks

export GOPATH=~/socks

cd socks/src/socks

go build

./socks


在运行前，需要写好自己的配置文件，配置文件名为**socks.config**，数据格式为json，内容类似如下：
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
			        "cryptoMethod": "rc4",
			        "password": "abcd#1234",
			        "addr": "54.64.248.242:9999"
		        },
		        {
			        "cryptoMethod": "rc4",
			        "password": "abcd#1234",
			        "addr": "54.64.214.156:9999"
		        },
		        {
			        "cryptoMethod": "rc4",
			        "password": "abcd#1234",
			        "addr": "54.64.73.132:9999"
		        }
            ]
        }
    ]
}

```
pprof               - 用于观察程序的性能数据和栈状态，不需要可以不设置
configs             - 是一个CONFIG数组。可以配置多个CONFIG，每个CONFIG表示一个本地监听的端口以及对应的UPSTREAM列表
httpProxyAddr       - 表示本地HTTP RPOXY的监听地址和端口
socks4Addr          - 表示本地SOCKS4监听地址和端口
socks5Addr          - 表示本地SOCKS5监听地址和端口
localCryptoMethod   - 本地SOCKS5的加密算法，如果为空，表示不开启加密，否则就按指定的算法实现加密。可以配置"rc4"或者"des"
localPasssword      - 加密算法使用到的密钥
dnsCacheTimeout     - 表示DNS的缓存时间,单位为分钟
upstream            - 表示对应的上端服务列表
cryptoMethod        - 表示上端使用到的加密算法
password            - 表示上端加密算法的密钥
addr                - 表示对应的服务端地址和端口


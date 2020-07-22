## 内网穿透工具

- 一款内网穿透工具，目前只支持tcp协议，支持多链接同时访问。
- 编译make
- 运行
	- 在公网服务器上运行proxyserver.cd bin & ./start_server.sh
	- 在内网服务器上运行proxyclient.cd bin & ./start_client.sh
- config/server.yml
	- controladdr:公网服务器对内网服务器开放的控制链路的端口
	- clientaddr:公网服务器对外网服务器开放的服务的端口，可以配置多个
	- useraddr:公网服务器对用户开放的服务器端口，可以配置多个
- config/client.yml
	- controladdr:内网服务器访问的公网服务器的的控制链路的端口，和server你的配置一直
	- proxys:转发的控制，remote代表对公网的端口连接(和server的clientaddr一致)，local代表的转发给内网的服务器的端口连接
- 注意配置转发多个服务器的时候顺序要一致


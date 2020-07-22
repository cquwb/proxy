package config

const (
	ServerTypeUser = iota
	ServerTypeClient
	ServerTypeClientControl
)

const (
	ClientTypeServer = iota
	ClientTypeControl
	ClientTypeDst //目标服务的地址
)

//控制消息类型
const (
	ControlNewConn = iota //通知内网客户端建立一个新连接
)

const (
	MsgKeepAlive = iota + 10001
	MsgNewConn
)

const (
	ConnectNil int32 = iota //未连接
	Connecting              //连接中
	Connected               //已连接
)

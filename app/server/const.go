package main

const (
	ServerTypeUser = iota
	ServerTypeClient
	ServerTypeClientControl
)

//控制消息类型
const (
	ControlNewConn = iota //通知内网客户端建立一个新连接
)

const (
	MsgKeepAlive = iota + 10001
	MsgNewConn
)

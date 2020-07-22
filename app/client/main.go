package main

import (
	"app/config"
	"app/tcp"
	"context"
	"flag"
	"fmt"
	"github.com/jinzhu/configor"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	l4g "github.com/ivanabc/log4go"
)

/**
todo 需要关闭一个内网的键值对，释放资源

*/
var gConfig = &config.Client{}

var configFile = flag.String("config", "../config/client.yml", "配置文件")

var gControlConnState = config.ConnectNil
var gCtx = context.Background()

func main() {
	err := configor.New(&configor.Config{Debug: true}).Load(gConfig, *configFile)
	if err != nil {
		panic(fmt.Sprintf("load config error %s", err))
	}
	serverCtx, f := context.WithCancel(gCtx)
	defer f()
	go ConnectControlServer(serverCtx)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	timer := time.NewTicker(10 * time.Second)
QUIT:
	for {
		select {
		case <-sigs:
			l4g.Debug("close \n")
			break QUIT
		case <-timer.C:
			ReconnectControlServer()
		}
	}
}

func ConnectControlServer(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&gControlConnState, config.ConnectNil, config.Connecting) {
		return
	}
	addr := gConfig.ControlAddr
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		SetStateNil()
		l4g.Error("connect control server error %s %s", addr, err)
		return
	}
	SetStateSuccess()
	l4g.Info("connect control server success")
	handleControlConn(ctx, conn)
}

func ConnectRemoteServer(k int) net.Conn {
	addr := gConfig.Proxys[k].Remote
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		l4g.Error("connect server error %s %s", addr, err)
		return nil
	}
	return conn
}

func ConnectLocalServer(k int) net.Conn {
	addr := gConfig.Proxys[k].Local
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		l4g.Error("connect server error %s %s", addr, err)
		return nil
	}
	return conn
}

func handleControlConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	defer SetStateNil()
	for {
		message, err := tcp.ReadMsg(conn)
		if err != nil {
			l4g.Debug("handleControplConn error %s", err)
			return
		}
		handleMessage(message)
	}
}

func SetStateNil() {
	atomic.StoreInt32(&gControlConnState, config.ConnectNil)
}

func SetStateSuccess() {
	atomic.StoreInt32(&gControlConnState, config.Connected)
}

func handleMessage(message *tcp.Message) {
	if message.GetT() == config.MsgNewConn {
		CreateNewPipe(message.GetK())
	}
}

func ReconnectControlServer() {
	if atomic.LoadInt32(&gControlConnState) == config.ConnectNil {
		l4g.Info("begin reconnect control server")
		ctx, _ := context.WithCancel(gCtx)
		go ConnectControlServer(ctx)
	}
}

func CreateNewPipe(k int) {
	l4g.Debug("createNewPipe %d", k)
	conn1 := ConnectRemoteServer(k)
	if conn1 == nil {
		l4g.Error("createNewPipe connect remote server error %d", k)
		return
	}
	conn2 := ConnectLocalServer(k)
	if conn2 == nil {
		l4g.Error("createNewPipe connect remote server error %d", k)
		return
	}
	tcp.SwapConn(conn1, conn2)

}

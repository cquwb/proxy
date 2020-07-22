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
	"sync"
	"syscall"
	"time"

	l4g "github.com/ivanabc/log4go"
)

var gConfig = &config.Server{}

var configFile = flag.String("config", "../config/server.yml", "配置文件")

var gClientConnChans = make([]chan net.Conn, 0, 100) //内网客户端新建的链接
var gControlMsg = make(chan *tcp.Message, 10)

var gListens = make([]net.Listener, 0, 100)

var gMutex sync.Mutex

func main() {
	bg := context.Background()
	serverCtx, f := context.WithCancel(bg)
	defer f()
	err := configor.New(&configor.Config{Debug: true}).Load(gConfig, *configFile)
	if err != nil {
		panic(fmt.Sprintf("load config error %s", err))
	}
	for k, v := range gConfig.UserAddrs {
		ctx, _ := context.WithCancel(serverCtx)
		go BeginUserServer(ctx, k, v)
		gClientConnChans = append(gClientConnChans, make(chan net.Conn, 100))
	}
	ctx, _ := context.WithCancel(serverCtx)
	go BeginControlServer(ctx)
	for k, v := range gConfig.ClientAddrs {
		ctx, _ := context.WithCancel(serverCtx)
		go BeginClientServer(ctx, k, v)
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
QUIT:
	for {
		select {
		case <-sigs:
			l4g.Debug("close \n")
			break QUIT
		}
	}
	closeAllListen()
}

func BeginUserServer(ctx context.Context, k int, addr string) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(fmt.Sprintf("listen server error %d %s", k, addr))
	}
	defer l.Close()
	l4g.Debug("begin listen server %s", addr)
	for {
		conn, err := l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				l4g.Error("tcp accept error %s", err)
				continue
			} else {
				l4g.Error("tcp accept big error %s", err)
				return
			}
		}
		go HandleUserConn(ctx, k, conn)
	}
}

func BeginControlServer(ctx context.Context) {
	addr := gConfig.ControlAddr
	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(fmt.Sprintf("listen control server error %s", addr))
	}
	defer l.Close()
	l4g.Debug("begin listen control server %s", addr)
	for {
		conn, err := l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				l4g.Error("tcp accept error %s", err)
				continue
			} else {
				l4g.Error("tcp accept big error %s", err)
				return
			}
		}
		go HandleControlConn(ctx, conn)
	}
}

func BeginClientServer(ctx context.Context, k int, addr string) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(fmt.Sprintf("listen server error %d %s", k, addr))
	}
	defer l.Close()
	l4g.Debug("begin listen server %s", addr)
	for {
		conn, err := l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				l4g.Error("tcp accept error %s", err)
				continue
			} else {
				l4g.Error("tcp accept big error %s", err)
				return
			}
		}
		go HandleClientConn(ctx, k, conn)
	}
}

//todo 这里怎么关闭资源
func HandleUserConn(ctx context.Context, k int, conn net.Conn) {
	l4g.Debug("user server accept %d", k)
	message := tcp.NewMessage(config.MsgNewConn, k)
	gControlMsg <- message
	select {
	case dstConn := <-gClientConnChans[k]:
		tcp.SwapConn(conn, dstConn)
	case <-time.After(5 * time.Second):
		conn.Close() //关闭连接
		l4g.Error("HandlerUserConn %d get new conn time out", k)
	}
}

//这个conn重联了必须关闭掉，防止重连之后老的还在
func HandleControlConn(ctx context.Context, conn net.Conn) {
	l4g.Info("new Control conn in")
	timer := time.NewTicker(10 * time.Second)
	defer timer.Stop()
	defer l4g.Info("old control conn finish")
	for {
		select {
		case message := <-gControlMsg:
			err := tcp.SendMsg(conn, message)
			if err != nil {
				l4g.Error("send msg to control client error %s", err)
				return
			}
		case <-timer.C:
			message := tcp.NewMessage(config.MsgKeepAlive, 0)
			err := tcp.SendMsg(conn, message)
			if err != nil {
				l4g.Error("send msg to control client error %s", err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func HandleClientConn(ctx context.Context, k int, conn net.Conn) {
	l4g.Debug("client server accept %d %+v", k, conn)
	gClientConnChans[k] <- conn
}

func addListen(l net.Listener) {
	gMutex.Lock()
	defer gMutex.Unlock()
	gListens = append(gListens, l)
}

func closeAllListen() {
	gMutex.Lock()
	defer gMutex.Unlock()
	for _, v := range gListens {
		v.Close()
	}

}

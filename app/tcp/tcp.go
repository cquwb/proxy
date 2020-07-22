package tcp

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"time"

	l4g "github.com/ivanabc/log4go"
)

var ReadLenError = errors.New("read len error")

func SendMsg(conn net.Conn, message *Message) error {
	data := make([]byte, 8)
	binary.BigEndian.PutUint32(data[:4], uint32(message.t))
	binary.BigEndian.PutUint32(data[4:], uint32(message.k))
	_, err := conn.Write(data)
	return err
}

func ReadMsg(conn net.Conn) (*Message, error) {
	data := make([]byte, 8)
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if n, err := io.ReadFull(conn, data); err != nil {
		l4g.Error("ReadMsg error %s", err)
		return nil, err
	} else {
		if n != 8 {
			l4g.Error("ReadMsg len error %d", n)
			return nil, ReadLenError
		}
		message := &Message{}
		message.t = int(binary.BigEndian.Uint32(data[0:4]))
		message.k = int(binary.BigEndian.Uint32(data[4:8]))
		l4g.Debug("read msg is %v", message)
		return message, nil
	}
}

func SwapConn(srcConn, dstConn net.Conn) {
	go copyMsg(srcConn, dstConn)
	go copyMsg(dstConn, srcConn)
}

//注意这里只关闭这一个
func copyMsg(srcConn, dstConn net.Conn) {
	defer dstConn.Close()
	if n, err := io.Copy(dstConn, srcConn); err != nil {
		l4g.Error("copyMsg error %s %s %s %n", srcConn, dstConn, err, n)
	}
}

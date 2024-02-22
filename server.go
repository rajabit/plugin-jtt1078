package jtt1078

import (
	"bufio"
	"net"

	"go.uber.org/zap"
	. "m7s.live/engine/v4"
)

type NetStream struct {
	*NetConnection
	StreamID uint32
}

type AVSender struct {
	*Jtt1078Sender
	firstSent bool
}

type Jtt1078Receiver struct {
	Publisher
	NetStream
}

type Jtt1078Sender struct {
	Subscriber
	NetStream
}

type NetConnection struct {
	*bufio.Reader `json:"-" yaml:"-"`
	net.Conn      `json:"-" yaml:"-"`
	readSeqNum    uint32 // 当前读的字节
	writeSeqNum   uint32 // 当前写的字节
	totalWrite    uint32 // 总共写了多少字节
	totalRead     uint32 // 总共读了多少字节
}

func NewNetConnection(conn net.Conn) *NetConnection {
	return &NetConnection{
		Conn:   conn,
		Reader: bufio.NewReader(conn),
	}
}

func (config *Jtt1078Config) ServeTCP(conn net.Conn) {
	defer conn.Close()
	logger := Jtt1078Plugin.Logger.With(zap.String("remote", conn.RemoteAddr().String()))
	senders := make(map[uint32]*Jtt1078Subscriber)
	receivers := make(map[uint32]*Jtt1078Receiver)
	var err error
	logger.Info("conn")
	defer func() {
		ze := zap.Error(err)
		logger.Info("conn close", ze)
		for _, sender := range senders {
			sender.Stop(ze)
		}
		for _, receiver := range receivers {
			receiver.Stop(ze)
		}
	}()
	NewNetConnection(conn)
}

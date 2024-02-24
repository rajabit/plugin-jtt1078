package jtt1078

import (
	"bufio"
	"fmt"
	"net"
)

type NetConnection struct {
	*bufio.Reader `json:"-" yaml:"-"`
	net.Conn      `json:"-" yaml:"-"`
	readSeqNum    uint32 // 当前读的字节
	writeSeqNum   uint32 // 当前写的字节

}

func NewNetConnection(conn net.Conn) *NetConnection {
	return &NetConnection{
		Conn:   conn,
		Reader: bufio.NewReader(conn),
	}
}

func (conn *NetConnection) RecvJtt1078RTP() (pkg *Jtt1078RTP, err error) {
	for pkg == nil && err == nil {
		conn.readJtt1078RTP()
	}
	return
}

func (conn *NetConnection) readJtt1078RTP() (pkg *Jtt1078RTP, err error) {
	head, err := conn.ReadByte()
	if err != nil {
		return nil, err
	}
	fmt.Print(head)
	return
}

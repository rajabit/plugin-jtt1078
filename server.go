package jtt1078

import (
	"bufio"
	"fmt"
	"net"

	"go.uber.org/zap"
	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/util"
)

type NetStream struct {
	*NetConnection
}

type AVSender struct {
	*JTT1078Sender
}

type JTT1078Receiver struct {
	Publisher
}

type JTT1078Sender struct {
	Subscriber
}

type NetConnection struct {
	*bufio.Reader `json:"-" yaml:"-"`
	net.Conn      `json:"-" yaml:"-"`
}

func NewNetConnection(conn net.Conn) *NetConnection {
	return &NetConnection{
		Conn:   conn,
		Reader: bufio.NewReader(conn),
	}
}

// func (config *JTT1078Config) ServeTCP(conn net.Conn) {
// 	defer conn.Close()
// 	logger := JTT1078Plugin.Logger.With(zap.String("remote", conn.RemoteAddr().String()))
// 	senders := make(map[uint32]*JTT1078Subscriber)
// 	receivers := make(map[uint32]*JTT1078Receiver)
// 	var err error
// 	logger.Info("conn")
// 	defer func() {
// 		ze := zap.Error(err)
// 		logger.Info("conn close", ze)
// 		for _, sender := range senders {
// 			sender.Stop(ze)
// 		}
// 		for _, receiver := range receivers {
// 			receiver.Stop(ze)
// 		}
// 	}()
// 	nc := NewNetConnection(conn)
// 	for {
// 		if msg, err = nc.RecvMessage(); err == nil {
// 			if msg.MessageLength <= 0 {
// 				continue
// 			}
// 			switch msg.MessageTypeID {
// 			case RTMP_MSG_AMF0_COMMAND:
// 			}
// 		}
// 	}
// }

func (c *JTT1078Config) ServeTCP(conn net.Conn) {
	fmt.Println("JTT1078 ServeTCP...")
	reader := TCP1078RTP{
		Conn: conn,
	}

	reader.Start(func(data util.Buffer) (err error) {
		var rtpPacket Packet
		if err = rtpPacket.Unmarshal(data); err != nil {
			JTT1078Plugin.Error("JTT1078 decode rtp error:", zap.Error(err))
		}
		return
	})
}

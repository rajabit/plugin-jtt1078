package jtt1078

import (
	"bufio"
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

type JTT1078Stream struct {
	Flag bool
	*JTT1078Publisher
	net.Conn
}

func (c *JTT1078Config) ServeTCP(conn net.Conn) {
	JTT1078Plugin.Info("JTT1078Config ServeTCP")
	reader := TCP1078RTP{
		Conn: conn,
	}
	tcpAddr := zap.String("tcp", conn.LocalAddr().String())
	var puber *JTT1078Publisher
	// var devStream *JTT1078Stream

	err := reader.Start(func(data util.Buffer) (err error) {
		var jtt1078Pkg Packet
		if err = jtt1078Pkg.Unmarshal(data); err != nil {
			JTT1078Plugin.Error("JTT1078 decode rtp error:", zap.Error(err))
		}
		JTT1078Plugin.Info("receive jtt1078 stream @", tcpAddr)
		JTT1078Plugin.Info("SequenceNumber", zap.Uint16("sn", jtt1078Pkg.SequenceNumber))
		if puber == nil {
			puber = new(JTT1078Publisher)
			if JTT1078Plugin.Publish("live/"+jtt1078Pkg.getLiveAddr(), puber) == nil {
				//注册成功
				puber.Info("发布流注册成功...", zap.String("@", jtt1078Pkg.getLiveAddr()))
				puber.PushPS(&jtt1078Pkg)
				return
			}
		} else {
			puber.PushPS(&jtt1078Pkg)
		}
		return
	})
	if puber != nil {
		puber.Stop(zap.Error(err))
		puber.Info("stop receive stream from ", tcpAddr)
	}
}

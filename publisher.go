package jtt1078

import (
	"os"
	"time"

	. "m7s.live/engine/v4"
	. "m7s.live/engine/v4/track"
)

// 自定义发布者
type JTT1078Publisher struct {
	Publisher
	Packet      `json:"-" yaml:"-"`
	lastReceive time.Time
	dump        *os.File
	dumpLen     []byte
}

// 发布者事件回调
func (pub *JTT1078Publisher) OnEvent(event any) {
	switch v := event.(type) {
	case IPublisher: //代表发布成功事件
	case SEclose: //代表关闭事件
	case SEKick: //被踢出
	case ISubscriber:
		if v.IsClosed() {
			//订阅者离开
		} else {
			//订阅者进入
		}

	default:
		pub.Publisher.OnEvent(event)
	}
}

// func (p *JTT1078Publisher) ServeTCP(conn net.Conn) {
// 	reader := TCP1078RTP{
// 		Conn: conn,
// 	}
// 	p.SetIO(conn)
// 	defer p.Stop()
// 	tcpAddr := zap.String("tcp", conn.LocalAddr().String())
// 	p.Info("start receive ps stream from", tcpAddr)
// 	defer p.Info("stop receive ps stream from", tcpAddr)
// 	reader.Start(p.PushPS)
// }

// func (p *JTT1078Publisher) PushPS(ps util.Buffer) (err error) {
// 	if err = p.Unmarshal(ps); err != nil {
// 		p.Error("jtt1078 decode rtp error:", zap.Error(err))
// 	} else if !p.IsClosed() {
// 		p.writeDump(ps)
// 	}
// 	// p.pushPS()
// 	return
// }

// func (p *JTT1078Publisher) writeDump(ps util.Buffer) {
// 	if p.dump != nil {
// 		util.PutBE(p.dumpLen[:4], ps.Len())
// 		if p.lastReceive.IsZero() {
// 			util.PutBE(p.dumpLen[4:], 0)
// 		} else {
// 			util.PutBE(p.dumpLen[4:], uint16(time.Since(p.lastReceive).Milliseconds()))
// 		}
// 		p.lastReceive = time.Now()
// 		p.dump.Write(p.dumpLen)
// 		p.dump.Write(ps)
// 	}
// }

func (p *JTT1078Publisher) PushPS(pkt *Packet) (err error) {
	if p.PT == 98 {
		if p.VideoTrack == nil {
			p.VideoTrack = NewH264(p.Publisher.Stream)
		}
		//fmt.Println(p.Packet)
		p.WriteAnnexB(uint32(pkt.Timestamp), 0, pkt.Payload)
		//p.WriteSliceBytes(p.Packet.Payload)
	}
	return
}

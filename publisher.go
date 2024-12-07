package jtt1078

import (
	"os"
	"time"

	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/track"
	"m7s.live/engine/v4/util"
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
		pub.Publisher.OnEvent(event)
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
	// JTT1078Plugin.Logger.Info("PT", zap.Uint8("pt", pkt.PT))
	if pkt.PT == 98 || pkt.PT == 99 { // H264视频 || H265视频
		if p.VideoTrack == nil {
			p.VideoTrack = track.NewH264(p.Publisher.Stream.Publisher)
		}
		// JTT1078Plugin.Logger.Info("SequenceNumber", zap.Uint16("sn", pkt.SequenceNumber))
		// JTT1078Plugin.Logger.Info("Timestamp", zap.Uint64("ts", pkt.Timestamp))
		p.VideoTrack.WriteAnnexB(uint32(pkt.Timestamp)*90, uint32(pkt.Timestamp)*90, pkt.Payload)
	}
	if pkt.PT == 6 || pkt.PT == 7 { // G711A || G711U
		if p.AudioTrack == nil {
			if pkt.PT == 6 { // G711A
				p.AudioTrack = track.NewG711(p.Publisher.Stream.Publisher, true)
			} else { // G711U
				p.AudioTrack = track.NewG711(p.Publisher.Stream.Publisher, false)
			}
		}
		p.AudioTrack.WriteRawBytes(uint32(pkt.Timestamp)*90, util.ReuseBuffer{pkt.Payload})
	}
	return
}

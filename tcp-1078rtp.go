package jtt1078

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"m7s.live/engine/v4/util"
)

type TCP1078RTP struct {
	net.Conn
}

func (t *TCP1078RTP) Start(onRTP func(util.Buffer) error) (err error) {
	reader := bufio.NewReader(t.Conn)
	headerBuf := make(util.Buffer, 30)
	headerBuf_16 := make([]byte, 16)
	headerBuf_st := make([]byte, 8)   // Timestamp
	headerBuf_lfi := make([]byte, 4)  // LastIFrameInterval + LastFrameInterval
	headerBuf_plen := make([]byte, 2) // PayloadLen
	for err == nil {
		if _, err = io.ReadFull(reader, headerBuf_16); err != nil { // 读取头部的固定部分（16个字节）
			return
		}

		//1078RTP正常以0x30316364开头，如果不是，说明不对，此处并非开头，
		//而是可能之前的包不完整导致的，需要往后查找
		var hd_offset = 0
		for i := 0; i < len(headerBuf_16)-4; i++ { // 防止读溢出，所以要 "-4"
			if headerBuf[i] == 0x30 && headerBuf[i+1] == 0x31 && headerBuf[i+2] == 0x63 && headerBuf[i+3] == 0x64 {
				hd_offset = i
				// 一直找到 rtp 头为止
				break
			}
		}

		// 偏离了N个字节，就再读取N个字节，确保16个字节固定头部
		if hd_offset > 0 {
			if _, err = io.ReadFull(reader, headerBuf_16[hd_offset:]); err != nil {
				return
			}
		}
		copy(headerBuf, headerBuf_16) // 把前16字节先写入缓存

		var headerLen = 16
		var dtype_sflag uint8 = headerBuf_16[15]
		if (dtype_sflag & 0xf0) != 0x40 { // 数据类型不是0100，则有8字节的时间戳
			// 读取header的Timestamp (8个字节)
			if _, err = io.ReadFull(reader, headerBuf_st); err != nil {
				return
			}
			copy(headerBuf[16:], headerBuf_st)
			headerLen = headerLen + 8
		}
		if (dtype_sflag&0xf0) == 0x00 || (dtype_sflag&0xf0) == 0x10 || (dtype_sflag&0xf0) == 0x20 {
			// 数据类型是视频帧，
			// 则读取header的Last I Frame Interval (WORD) 和 Last Frame Interval (WORD)
			if _, err = io.ReadFull(reader, headerBuf_lfi); err != nil {
				return
			}
			headerBuf = append(headerBuf, headerBuf_lfi[:]...)
			copy(headerBuf[headerLen:], headerBuf_lfi)
			headerLen = headerLen + 4
		}

		// 读取header的PayloadLen（2个字节）
		if _, err = io.ReadFull(reader, headerBuf_plen); err != nil {
			return
		}
		copy(headerBuf[headerLen:], headerBuf_plen)
		headerBuf = append(headerBuf, headerBuf_plen[:]...)
		payloadLen := binary.BigEndian.Uint16(headerBuf_plen[0:2])
		headerLen = headerLen + 2
		fmt.Printf(".h.%d+", headerLen)
		fmt.Printf(".p.%d\n", payloadLen)

		buffer := make(util.Buffer, headerLen+(int)(payloadLen))
		copy(buffer, headerBuf)

		if _, err = io.ReadFull(reader, buffer[headerLen:]); err != nil {
			return
		}
		// fmt.Printf(".h4.%x\n", buffer[0:4])
		// fmt.Printf(".t4.%x\n", buffer[608:612])

		err = onRTP(buffer)
	}
	return
}

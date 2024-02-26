package jtt1078

import (
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
)

// Header represents an RTP packet header
type Header struct {
	FixedFlag          uint32
	V_P_X_CC           uint8
	M                  bool
	PT                 uint8
	SequenceNumber     uint16
	SIM                [6]byte // BCD编码
	LogicChannel       uint8
	Datatype_Splitflag uint8
	Timestamp          uint64
	LastIFrameInterval uint16
	LastFrameInterval  uint16
	PayloadLen         uint16
	// _pub_addr          uint64  // 00 SIM[6] LogicChannel，一共8个字节
}

// Packet represents an RTP Packet
type Packet struct {
	Header
	Payload []byte
}

type word [2]byte
type dword [4]byte
type qword [8]byte

// Decoder is used to decode BCD converted bytes into decimal string.
//
// Decoder may be copied with no side effects.
type Decoder struct {
	// if the input contains filler nibble in the middle, default
	// behaviour is to treat this as an error. You can tell decoder to
	// resume decoding quietly in that case by setting this.
	IgnoreFiller bool

	// two nibbles (1 byte) to 2 symbols mapping; example: 0x45 ->
	// '45' or '54' depending on nibble swapping additional 2 bytes of
	// dword should be 0, otherwise given byte is unacceptable
	hashWord [0x100]dword

	// one finishing byte with filler nibble to 1 symbol mapping;
	// example: 0x4f -> '4' (filler=0xf, swap=false)
	// additional byte of word should 0, otherise given nibble is
	// unacceptable
	hashByte [0x100]word
}

// String helps with debugging by printing packet information in a readable way
func (p Packet) String() string {
	out := "JTT1078RTP PACKET:\n"

	out += fmt.Sprintf("\tM: %v\n", p.M)
	out += fmt.Sprintf("\tPT: %d\n", p.PT)
	out += fmt.Sprintf("\tSequence Number: %d\n", p.SequenceNumber)
	out += fmt.Sprintf("\tSIM BCD6: %v\n", p.SIM)
	out += fmt.Sprintf("\tLogicChannel: %d\n", p.LogicChannel)
	out += fmt.Sprintf("\tTimestamp: %d\n", p.Timestamp)
	out += fmt.Sprintf("\tPayload Length: %d\n", p.PayloadLen)

	return out
}

func (p Packet) decodeSIM2String() string {
	// decode SIM BCD[6]
	dec := NewDecoder(Telephony)
	dst := make([]byte, DecodedLen(len(p.SIM)))
	var sim6 []byte = p.SIM[:]
	n, err := dec.Decode(dst, sim6)
	if err != nil {
		fmt.Println("bcd decode error")
		return ""
	}
	return string(dst[:n])
}

func (p Packet) getLiveAddr() string {
	// decode SIM BCD[6]
	dec := NewDecoder(Telephony)
	dst := make([]byte, DecodedLen(len(p.SIM)))
	var sim6 []byte = p.SIM[:]
	n, err := dec.Decode(dst, sim6)
	if err != nil {
		fmt.Println("bcd decode error")
		return ""
	}
	return string(dst[:n]) + "_" + strconv.Itoa(int(p.LogicChannel))
}

// Unmarshal parses the passed byte slice and stores the result in the Header.
// It returns the number of bytes read n and any error.
func (h *Header) Unmarshal(buf []byte) (n int, err error) { //nolint:gocognit
	h.FixedFlag = binary.BigEndian.Uint32(buf[0:4])
	if h.FixedFlag != 0x30316364 {
		return 0, fmt.Errorf("%x should be 0x30316364", h.FixedFlag)
	}

	h.V_P_X_CC = buf[4]
	h.M = (buf[5] >> 7 & 0x01) > 0
	h.PT = buf[5] & 0x7f
	h.SequenceNumber = binary.BigEndian.Uint16(buf[6:8])
	// fmt.Println(h.V_P_X_CC, " ", h.M, " ", h.PT, " ", h.SequenceNumber)
	for i := 8; i < 14; i++ {
		h.SIM[i-8] = buf[i]
	}

	h.LogicChannel = buf[14]
	h.Datatype_Splitflag = buf[15]
	var offset int = 0
	if (h.Datatype_Splitflag & 0xf0) != 0x40 { // 数据类型不是0100，则有8字节的时间戳
		h.Timestamp = binary.BigEndian.Uint64(buf[16:24])
		offset = offset + 8
	}

	if (h.Datatype_Splitflag&0xf0) == 0x00 || (h.Datatype_Splitflag&0xf0) == 0x10 || (h.Datatype_Splitflag&0xf0) == 0x20 {
		// 数据类型是视频帧，则有Last I Frame Interval (WORD) 和 Last Frame Interval (WORD)
		h.LastIFrameInterval = binary.BigEndian.Uint16(buf[16+offset : 16+offset+2])
		h.LastFrameInterval = binary.BigEndian.Uint16(buf[18+offset : 18+offset+2])
		offset = offset + 4
	}

	h.PayloadLen = binary.BigEndian.Uint16(buf[16+offset : 16+offset+2])

	return offset + 18, nil
}

// NewDecoder creates new Decoder from BCD configuration. If the
// configuration is invalid NewDecoder will panic.
func NewDecoder(config *BCD) *Decoder {
	if !checkBCD(config) {
		panic("BCD table is incorrect")
	}

	return &Decoder{
		hashWord: newHashDecWord(config),
		hashByte: newHashDecByte(config)}
}

// DecodedLen tells how much space is needed to store decoded string.
// Please note that it returns the max amount of possibly needed space
// because last octet may contain only one encoded digit. In that
// case the decoded length will be less by 1. For example, 4 octets
// may encode 7 or 8 digits.  Please examine the result of Decode to
// obtain the real value.
func DecodedLen(x int) int {
	return 2 * x
}

// Unmarshal parses the passed byte slice and stores the result in the Packet.
func (p *Packet) Unmarshal(buf []byte) error {
	n, err := p.Header.Unmarshal(buf)
	if err != nil {
		return err
	}

	end := n + int(p.Header.PayloadLen)
	p.Payload = buf[n:end]

	return nil
}

// Marshal serializes the header into bytes.
func (h Header) Marshal() (buf []byte, err error) {
	buf = make([]byte, h.MarshalSize())

	n, err := h.MarshalTo(buf)
	if err != nil {
		return nil, err
	}

	return buf[:n], nil
}

// MarshalTo serializes the header and writes to the buffer.
func (h Header) MarshalTo(buf []byte) (n int, err error) {
	size := h.MarshalSize()
	if size > len(buf) {
		return 0, io.ErrShortBuffer
	}

	binary.BigEndian.PutUint32(buf[0:4], h.FixedFlag)
	buf[4] = h.V_P_X_CC
	buf[5] = h.PT
	if h.M {
		buf[5] = h.PT | 0x80
	}
	binary.BigEndian.PutUint16(buf[6:8], h.SequenceNumber)
	for i := 8; i < 14; i++ {
		buf[i] = h.SIM[i-8]
	}
	buf[14] = h.LogicChannel

	// var paddr = make([]byte, 8)
	// for i := 1; i <= 6; i++ {
	// 	paddr[i] = h.SIM[i-1]
	// }
	// paddr[0] = 0x00
	// paddr[7] = buf[14]
	// binary.BigEndian.PutUint64(paddr, h._pub_addr)

	buf[15] = h.Datatype_Splitflag
	var offset int = 0
	if (h.Datatype_Splitflag & 0xf0) != 0x40 { // 数据类型不是0100，则有8字节的时间戳
		binary.BigEndian.PutUint64(buf[16:24], h.Timestamp)
		offset = offset + 8
	}

	if (h.Datatype_Splitflag&0xf0) == 0x00 || (h.Datatype_Splitflag&0xf0) == 0x10 || (h.Datatype_Splitflag&0xf0) == 0x20 {
		// 数据类型是视频帧，则有Last I Frame Interval (WORD) 和 Last Frame Interval (WORD)
		binary.BigEndian.PutUint16(buf[16+offset:16+offset+2], h.LastIFrameInterval)
		binary.BigEndian.PutUint16(buf[18+offset:18+offset+2], h.LastFrameInterval)
		offset = offset + 4
	}

	binary.BigEndian.PutUint16(buf[16+offset:16+offset+2], h.PayloadLen)

	return offset + 18, nil
}

// MarshalSize returns the size of the header once marshaled.
func (h Header) MarshalSize() int {
	size := 30
	if (h.Datatype_Splitflag & 0xf0) == 0x40 { // 数据类型不是0100，则有8字节的时间戳
		size = size - 8
	}

	if (h.Datatype_Splitflag&0xf0) != 0x00 && (h.Datatype_Splitflag&0xf0) != 0x10 && (h.Datatype_Splitflag&0xf0) == 0x20 {
		// 数据类型是视频帧，则有Last I Frame Interval (WORD) 和 Last Frame Interval (WORD)
		size = size - 4
	}

	return size
}

// Marshal serializes the packet into bytes.
func (p Packet) Marshal() (buf []byte, err error) {
	buf = make([]byte, p.MarshalSize())

	n, err := p.MarshalTo(buf)
	if err != nil {
		return nil, err
	}

	return buf[:n], nil
}

// MarshalTo serializes the packet and writes to the buffer.
func (p *Packet) MarshalTo(buf []byte) (n int, err error) {
	n, err = p.Header.MarshalTo(buf)
	if err != nil {
		return 0, err
	}

	// Make sure the buffer is large enough to hold the packet.
	if n+int(p.PayloadLen) > len(buf) {
		return 0, io.ErrShortBuffer
	}

	m := copy(buf[n:], p.Payload)

	return n + m, nil
}

// MarshalSize returns the size of the packet once marshaled.
func (p Packet) MarshalSize() int {
	return p.Header.MarshalSize() + len(p.Payload)
}

func (dec *Decoder) unpack(w []byte, b byte) (n int, end bool, err error) {
	if dw := dec.hashWord[b]; dw[2] == 0 {
		return copy(w, dw[:2]), false, nil
	}
	if dw := dec.hashByte[b]; dw[1] == 0 {
		return copy(w, dw[:1]), true, nil
	}
	return 0, false, ErrBadBCD
}

// Clone returns a deep copy of p.
func (p Packet) Clone() *Packet {
	clone := &Packet{}
	clone.Header = p.Header.Clone()
	if p.Payload != nil {
		clone.Payload = make([]byte, len(p.Payload))
		copy(clone.Payload, p.Payload)
	}
	return clone
}

// Clone returns a deep copy h.
func (h Header) Clone() Header {
	clone := h
	return clone
}

func newHashDecWord(config *BCD) (res [0x100]dword) {
	var w dword
	var b byte
	for i, _ := range res {
		// invalidating all bytes by default
		res[i] = dword{0xff, 0xff, 0xff, 0xff}
	}

	for c1, nib1 := range config.Map {
		for c2, nib2 := range config.Map {
			b = (nib1 << 4) + nib2&0xf
			if config.SwapNibbles {
				w = dword{c2, c1, 0, 0}
			} else {
				w = dword{c1, c2, 0, 0}
			}
			res[b] = w
		}
	}
	return
}

func newHashDecByte(config *BCD) (res [0x100]word) {
	var b byte
	for i, _ := range res {
		// invalidating all nibbles by default
		res[i] = word{0xff, 0xff}
	}
	for c, nib := range config.Map {
		if config.SwapNibbles {
			b = (config.Filler << 4) + nib&0xf
		} else {
			b = (nib << 4) + config.Filler&0xf
		}
		res[b] = word{c, 0}
	}
	return
}

// Decode parses BCD encoded bytes from src and tries to decode them
// to dst. Number of decoded bytes and possible error is returned.
func (dec *Decoder) Decode(dst, src []byte) (n int, err error) {
	if len(src) == 0 {
		return 0, nil
	}

	for _, c := range src[:len(src)-1] {
		wid, end, err := dec.unpack(dst[n:], c)
		switch {
		case err != nil: // invalid input
			return n, err
		case wid == 0: // no place in dst
			return n, nil
		case end && !dec.IgnoreFiller: // unexpected filler
			return n, ErrBadBCD
		}
		n += wid
	}

	c := src[len(src)-1]
	wid, _, err := dec.unpack(dst[n:], c)
	switch {
	case err != nil: // invalid input
		return n, err
	case wid == 0: // no place in dst
		return n, nil
	}
	n += wid
	return n, nil
}

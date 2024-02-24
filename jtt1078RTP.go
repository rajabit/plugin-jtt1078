package jtt1078

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
}

// Packet represents an RTP Packet
type Jtt1078RTP struct {
	Header
	Payload []byte
}

package jtt1078

import (
	"fmt"
	"net"

	"go.uber.org/zap"
)

func (config *Jtt1078Config) ServeTCP(conn net.Conn) {
	defer conn.Close()
	logger := Jtt1078Plugin.Logger.With(zap.String("remote", conn.RemoteAddr().String()))
	var err error
	logger.Info("conn")
	defer func() {
		ze := zap.Error(err)
		logger.Info("conn close", ze)
	}()
	nc := NewNetConnection(conn)
	var pkg *Jtt1078RTP
	for {
		pkg, err = nc.RecvJtt1078RTP()
		fmt.Print(pkg)
		logger.Info("rtmp client closed")
		return
	}
}

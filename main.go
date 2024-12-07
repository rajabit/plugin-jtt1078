package jtt1078

import (
	"net/http"

	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/config"
	"m7s.live/engine/v4/track"
)

// JTT1078Config /*
type JTT1078Config struct {
	config.HTTP
	config.Publish
	config.Subscribe
	config.TCP
}

var conf = &JTT1078Config{
	TCP: config.TCP{ListenAddr: ":7222"},
}

// Install the plugin
var JTT1078Plugin = InstallPlugin(conf)

// Plugin event callback from the event bus
func (c *JTT1078Config) OnEvent(event any) {
	switch event.(type) {
	case FirstConfig: // 插件启动事件
		break
	}
}

// http://localhost:8080/JTT1078/api/test/pub
func (conf *JTT1078Config) API_test_pub(rw http.ResponseWriter, r *http.Request) {
	var pub JTT1078Publisher
	err := JTT1078Plugin.Publish("JTT1078/test", &pub)
	if err != nil {
		rw.Write([]byte(err.Error()))
		return
	} else {
		vt := track.NewH264(pub.Stream.Publisher)
		// 根据实际情况写入视频帧，需要注意pts和dts需要写入正确的值 即毫秒数*90
		vt.WriteAnnexB(0, 0, []byte{0, 0, 0, 1})
	}
	rw.Write([]byte("test_pub"))
}

// http://localhost:8080/JTT1078/api/test/sub
func (conf *JTT1078Config) API_test_sub(rw http.ResponseWriter, r *http.Request) {
	var sub JTT1078Subscriber
	err := JTT1078Plugin.Subscribe("JTT1078/test", &sub)
	if err != nil {
		rw.Write([]byte(err.Error()))
		return
	} else {
		sub.PlayRaw()
	}
	rw.Write([]byte("test_sub"))
}

// 自定义订阅者
type JTT1078Subscriber struct {
	Subscriber
}

// 订阅者事件回调
func (sub *JTT1078Subscriber) OnEvent(event any) {
	switch event.(type) {
	case ISubscriber:
		// 订阅成功
	case AudioFrame:
		// 音频帧处理
	case VideoFrame:
		// 视频帧处理
	default:
		sub.Subscriber.OnEvent(event)
	}
}

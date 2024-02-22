package jtt1078

import (
	"net/http"

	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/config"
	"m7s.live/engine/v4/track"
)

/*
自定义配置结构体
配置文件中可以添加相关配置来设置结构体的值
jtt1078:

	http:
	publish:
	subscribe:
	foo: bar
*/
type Jtt1078Config struct {
	config.HTTP
	config.Publish
	config.Subscribe
	config.TCP
}

var jtt1078Config = &Jtt1078Config{
	TCP: config.TCP{ListenAddr: ":9024"},
}

// 安装插件
var Jtt1078Plugin = InstallPlugin(jtt1078Config)

// 插件事件回调，来自事件总线
func (c *Jtt1078Config) OnEvent(event any) {
	switch event.(type) {
	case FirstConfig: // 插件启动事件
		if c.TCP.ListenAddr != "" {
			go c.ListenTCP(Jtt1078Plugin, c)
		}
		break
	}
}

// http://localhost:8080/jtt1078/api/test/pub
func (conf *Jtt1078Config) API_test_pub(rw http.ResponseWriter, r *http.Request) {
	var pub Jtt1078Publisher
	err := Jtt1078Plugin.Publish("jtt1078/test", &pub)
	if err != nil {
		rw.Write([]byte(err.Error()))
		return
	} else {
		vt := track.NewH264(pub.Stream)
		// 根据实际情况写入视频帧，需要注意pts和dts需要写入正确的值 即毫秒数*90
		vt.WriteAnnexB(0, 0, []byte{0, 0, 0, 1})
	}
	rw.Write([]byte("test_pub"))
}

// http://localhost:8080/jtt1078/api/test/sub
func (conf *Jtt1078Config) API_test_sub(rw http.ResponseWriter, r *http.Request) {
	var sub Jtt1078Subscriber
	err := Jtt1078Plugin.Subscribe("jtt1078/test", &sub)
	if err != nil {
		rw.Write([]byte(err.Error()))
		return
	} else {
		sub.PlayRaw()
	}
	rw.Write([]byte("test_sub"))
}

// 自定义发布者
type Jtt1078Publisher struct {
	Publisher
}

// 发布者事件回调
func (pub *Jtt1078Publisher) OnEvent(event any) {
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

// 自定义订阅者
type Jtt1078Subscriber struct {
	Subscriber
}

// 订阅者事件回调
func (sub *Jtt1078Subscriber) OnEvent(event any) {
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

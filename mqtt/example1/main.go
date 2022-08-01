package main

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"sync"
	"time"
)

// 连接的回调函数
var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("已成功连接")
}

// 丢失连接的回调函数
var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect loss: %v \n", err)
}

// 创建全局 mqtt sub 消息处理 handler
var messageSubHandler mqtt.MessageHandler = func(client mqtt.Client, message mqtt.Message) {
	fmt.Println("收到订阅消息")
	fmt.Printf("Sub Client Topic: %s \n", message.Topic())
	fmt.Printf("Sub Client msg: %s \n", message.Payload())
}

var (
	client mqtt.Client
	wg     sync.WaitGroup
)

func MqttConnect() {
	opts := mqtt.NewClientOptions().AddBroker("tcp://175.178.97.141:1883").SetClientID("emqx_test_client66666")
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(1 * time.Second)

	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	//连接
	client = mqtt.NewClient(opts)
	//客户端连接判断
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}

func mqttSubScribe(topic string, qos byte) {
	defer wg.Done()
	for {
		token := client.Subscribe(topic, qos, messageSubHandler)
		token.Wait()
	}
}

func testPublish(topic string, qos byte, retained bool, payload interface{}) {
	defer wg.Done()
	for {
		client.Publish(topic, qos, retained, payload)
		time.Sleep(time.Duration(3) * time.Second)
	}
}
func main() {
	//连接MQTT服务器
	MqttConnect()
	defer client.Disconnect(250) //注册销毁
	wg.Add(1)
	go mqttSubScribe("topic/test", 1)
	wg.Add(1)
	go testPublish("topic/test", 1, false, "test")
	wg.Wait()
}

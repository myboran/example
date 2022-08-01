package mq

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"time"
)

var client mqtt.Client

// 创建全局 mqtt publish 消息处理 handler
var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, message mqtt.Message) {
	fmt.Println("Push Message")
	fmt.Printf("TOPIC: %s\n", message.Topic())
	fmt.Printf("MSG: %s\n", message.Payload())
}

// 创建全局 mqtt sub 消息处理 handler
var messageSubHandler mqtt.MessageHandler = func(client mqtt.Client, message mqtt.Message) {
	fmt.Println("收到订阅消息")
	fmt.Printf("Sub Client Topic: %s \n", message.Topic())
	fmt.Printf("Sub Client msg: %s \n", message.Payload())
}

// 连接的回调函数
var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("新的连接: Connected")
}

// 丢失连接的回调函数
var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect loss: %v \n", err)
}

func MqttConnect(id string) {
	//配置
	opts := mqtt.NewClientOptions().AddBroker("tcp://175.178.97.141:1883").SetClientID(id)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetDefaultPublishHandler(messagePubHandler)
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

func Push(topic string, qos byte, retain bool, payload string) {
	fmt.Println("---------------------Push---------------------")
	// qos是服务质量: ==1: 一次, >=1: 至少一次, <=1:最多一次
	// retained: 表示mqtt服务器要保留这次推送的信息，如果有新的订阅者出现，就会把这消息推送给它（持久化推送）
	token := client.Publish(topic, qos, retain, payload)
	token.Wait()
	fmt.Printf("Push 发送数据: %v 至 topic: %v size: %v\n", payload, topic, len(payload))
	fmt.Println("Disconnect with broker")
}

func Subscription(topic string, qos byte, callback mqtt.MessageHandler) {
	fmt.Println("---------------------Subscription---------------------")
	if token := client.Subscribe(topic, qos, callback); token.Wait() && token.Error() != nil {
		fmt.Println("订阅消息失败: ", token.Error())
		return
	}
	fmt.Print("Subscribe topic " + topic + " success\n")
}

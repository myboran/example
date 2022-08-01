package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

const MqUrl = "amqp://center:123qwe@175.178.97.141:5672/center"

type RabbitMQ struct {
	exchange   string
	queue      string
	routingKey string
	conn       *amqp.Connection
	channel    *amqp.Channel
}

func NewRabbitMQ() *RabbitMQ {
	rmq := &RabbitMQ{}
	rmq.exchange = "x-delayed-message"
	rmq.queue = "delay_queue"
	rmq.routingKey = "log_delay"
	var err error
	rmq.conn, err = amqp.Dial(MqUrl)
	rmq.failOnErr(err, "创建连接错误")
	rmq.channel, err = rmq.conn.Channel()
	rmq.failOnErr(err, "获取channel失败")
	// 申请交换机
	err = rmq.channel.ExchangeDeclare(rmq.exchange, rmq.exchange, true, false, false, false, amqp.Table{
		"x-delayed-type": "direct",
	})
	rmq.failOnErr(err, "交换机申请失败")
	_, err = rmq.channel.QueueDeclare(rmq.queue, true, false, false, false, nil)
	rmq.failOnErr(err, "绑定交换机失败")
	return rmq
}

//错误处理函数
func (r *RabbitMQ) failOnErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s:%s", message, err)
	}
}

func (r *RabbitMQ) Destroy() {
	r.channel.Close()
	r.conn.Close()
}

// time is s
func (r *RabbitMQ) Pub(message string, time int64) {
	var err error
	err = r.channel.Publish(r.exchange, r.routingKey, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(message),
		Headers: amqp.Table{
			"x-delay": time * 1000,
		},
	})
	r.failOnErr(err, "Failed to publish a message")
}

func (r *RabbitMQ) Sub(name string) {
	msgCh, err := r.channel.Consume(r.queue, "sdas", false, false, false, false, nil)
	r.failOnErr(err, "Failed to register a consumer")
	for d := range msgCh {

		_ = d.Ack(true)
		log.Printf("%s 接收数据:%s", name, d.Body)
	}
	fmt.Println("退出")
}

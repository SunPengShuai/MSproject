package mqApi

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

// RabbitMQApi 实现 MqApi 接口
type RabbitMQApi struct {
	conn          *amqp.Connection
	channel       *amqp.Channel
	routingQueues map[string]amqp.Queue
	publicQueues  map[string]amqp.Queue
	exchange      string // 交换机名称
	exchangeType  string // 交换机类型
}

// NewRabbitMQApi 创建新的 RabbitMQ API 实例
// amqpURL：RabbitMQ连接URL
// queueName：队列名称
// exchange：交换机名称
// exchangeType：交换机类型 (fanout, direct, topic)
// routingKey：路由键（用于 direct 和 topic 模式）
func NewRabbitMQApi(amqpURL, exchange, exchangeType string) (*RabbitMQApi, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %v", err)
	}

	// 声明交换机
	err = channel.ExchangeDeclare(
		exchange,     // 交换机名称
		exchangeType, // 交换机类型 (fanout, direct, topic)
		true,         // 是否持久化
		false,        // 是否自动删除
		false,        // 是否阻塞
		false,        // 附加参数
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare an exchange: %v", err)
	}
	return &RabbitMQApi{
		conn:          conn,
		channel:       channel,
		exchange:      exchange,
		exchangeType:  exchangeType,
		publicQueues:  make(map[string]amqp.Queue),
		routingQueues: make(map[string]amqp.Queue),
	}, nil
}
func (r *RabbitMQApi) bindPubQ(cname string) error {
	q, ex := r.publicQueues[cname]
	if !ex {
		q, _ = r.channel.QueueDeclare(cname, true, false, false, false, nil)
		r.publicQueues[cname] = q
	}
	// 对于 fanout 类型交换机，路由键不需要设置
	err := r.channel.QueueBind(
		q.Name,     // 队列名称
		"",         // 路由键
		r.exchange, // 交换机名称
		false,      // 是否阻塞
		nil,        // 附加参数
	)
	if err != nil {
		return fmt.Errorf("failed to bind a queue: %v", err)
	}
	return nil
}
func (r *RabbitMQApi) bindRoutingQ(qname string, routingKey string) error {
	if _, ex := r.routingQueues[qname]; !ex {
		// 声明队列
		queue, err := r.channel.QueueDeclare(
			qname, // 队列名称
			true,  // 是否持久化
			false, // 是否自动删除
			false, // 是否独占
			false, // 是否阻塞
			nil,   // 额外属性
		)
		if err != nil {
			return fmt.Errorf("failed to bind a queue: %v", err)
		}
		r.routingQueues[qname] = queue
	}
	err := r.channel.QueueBind(
		qname,      // 队列名称
		routingKey, // 路由键
		r.exchange, // 交换机名称
		false,      // 是否阻塞
		nil,        // 附加参数
	)
	if err != nil {
		return fmt.Errorf("failed to bind a queue: %v", err)
	}
	return nil
}
func (r *RabbitMQApi) BindQ(qname string, routingKey string) error {
	if r.exchangeType == "funout" {
		err := r.bindPubQ(qname)
		if err != nil {
			return err
		}
	} else {
		err := r.bindRoutingQ(qname, routingKey)
		if err != nil {
			return err
		}
	}
	return nil
}
func (r *RabbitMQApi) recvSimple(qname string) (interface{}, error) {

	msgs, err := r.channel.Consume(
		r.routingQueues[qname].Name, // 队列名称
		"",                          // 消费者名称
		true,                        // 是否自动确认
		false,                       // 是否独占
		false,                       // 是否阻塞
		false,                       // 是否开启消息不丢失
		nil,                         // 附加参数
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register a consumer: %v", err)
	}

	for msg := range msgs {
		var mqMsg MqMsg
		if err := json.Unmarshal(msg.Body, &mqMsg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}
		return mqMsg, nil
	}

	return nil, fmt.Errorf("no message received")
}

func (r *RabbitMQApi) recvPublish(cname string) (interface{}, error) {

	msgs, err := r.channel.Consume(
		r.publicQueues[cname].Name, // 队列名称
		cname,                      // 消费者名称
		true,                       // 是否自动确认
		false,                      // 是否独占
		false,                      // 是否阻塞
		false,                      // 是否开启消息不丢失
		nil,                        // 附加参数
	)

	if err != nil {
		return nil, fmt.Errorf("failed to register a consumer: %v", err)
	}

	for msg := range msgs {
		var mqMsg MqMsg
		if err := json.Unmarshal(msg.Body, &mqMsg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}
		return mqMsg, nil
	}

	return nil, fmt.Errorf("no message received")
}

// RecvMsg 实现 MqApi 接口的 RecvMsg 方法
func (r *RabbitMQApi) RecvMsg(qname string) (interface{}, error) {
	if r.exchangeType == "funout" {
		msg, err := r.recvPublish(qname)
		if err != nil {
			return nil, err
		}
		return msg, nil
	} else {
		msg, err := r.recvSimple(qname)
		if err != nil {
			return nil, err
		}
		return msg, nil
	}
}

// SendMsg 实现 MqApi 接口的 SendMsg 方法
func (r *RabbitMQApi) SendMsg(msg MqMsg, routingKey string) error {
	// 将消息转换为 JSON 格式
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	// 发送消息到交换机
	err = r.channel.Publish(
		r.exchange, // 交换机名称
		routingKey, // 路由键
		false,      // 是否强制消息
		false,      // 是否强制消息
		amqp.Publishing{
			ContentType:  "application/json", // 消息格式
			Body:         body,               // 消息体
			DeliveryMode: amqp.Persistent,    // 设置消息持久化
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %v", err)
	}
	return nil
}

// Close 关闭 RabbitMQ 连接和通道
func (r *RabbitMQApi) Close() {
	r.channel.Close()
	r.conn.Close()
}

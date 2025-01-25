package test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"mqApi"
	"testing"
	"time"
)

// TestRabbitmq 测试 RabbitMQ API
func TestRabbitmqDirect(t *testing.T) {
	// RabbitMQ 服务地址，默认假设服务在本地运行
	amqpURL := "amqp://guest:guest@localhost:5672/"
	queueName := "testQueue"
	exchangeName := "testExchange"
	exchangeType := "direct" // 使用 direct 类型交换机
	routingKey := "testRoutingKey"
	// 创建 RabbitMQApi 实例
	rabbitMQApi, err := mqApi.NewRabbitMQApi(amqpURL, exchangeName, exchangeType)
	if err != nil {
		t.Fatalf("Failed to create RabbitMQ API: %v", err)
	}
	defer rabbitMQApi.Close() // 确保最后关闭连接
	rabbitMQApi.BindQ(queueName, routingKey)
	// 测试发送消息
	msg := mqApi.MqMsg{
		MsgType: mqApi.SimpleMsg,
		Data:    "Hello RabbitMQ!",
	}
	err = rabbitMQApi.SendMsg(msg, routingKey)
	if err != nil {
		t.Errorf("Failed to send message: %v", err)
	}

	// 测试接收消息
	// 使用 goroutine 来模拟异步接收消息
	go func() {
		time.Sleep(1 * time.Second) // 等待消息有时间发送到队列
		receivedMsg, err := rabbitMQApi.RecvMsg(queueName)
		fmt.Printf("Received message: %v\n", receivedMsg)
		if err != nil {
			t.Errorf("Failed to receive message: %v", err)
		}

		// 验证接收到的消息
		assert.NotNil(t, receivedMsg, "Received message should not be nil")
		receivedMqMsg, ok := receivedMsg.(mqApi.MqMsg)
		assert.True(t, ok, "Received message should be of type MqMsg")
		assert.Equal(t, msg.MsgType, receivedMqMsg.MsgType, "Message type should match")
		assert.Equal(t, msg.Data, receivedMqMsg.Data, "Message Data should match")
	}()

	// 确保接收操作能够在超时之前完成
	select {
	case <-time.After(2 * time.Second):
		t.Errorf("Timeout while waiting for the message")
	}
}

func TestRabbitmqFanout(t *testing.T) {
	// RabbitMQ 服务地址，默认假设服务在本地运行
	amqpURL := "amqp://guest:guest@localhost:5672/"
	queueNames := []string{"testQueue1", "testQueue2"}
	exchangeName := "testExchangeFanout"
	exchangeType := "fanout" // 使用 direct 类型交换机
	routingKey := ""
	// 创建 RabbitMQApi 实例
	rabbitMQApi, err := mqApi.NewRabbitMQApi(amqpURL, exchangeName, exchangeType)
	if err != nil {
		t.Fatalf("Failed to create RabbitMQ API: %v", err)
	}
	defer rabbitMQApi.Close() // 确保最后关闭连接
	for _, queueName := range queueNames {
		rabbitMQApi.BindQ(queueName, routingKey)
	}
	// 测试发送消息
	msg := mqApi.MqMsg{
		MsgType: mqApi.SimpleMsg,
		Data:    "Test Funout!",
	}
	err = rabbitMQApi.SendMsg(msg, routingKey)
	if err != nil {
		t.Errorf("Failed to send message: %v", err)
	}

	// 测试接收消息
	// 使用 goroutine 来模拟异步接收消息
	go func() {
		time.Sleep(1 * time.Second) // 等待消息有时间发送到队列
		for ind, queueName := range queueNames {
			receivedMsg, _ := rabbitMQApi.RecvMsg(queueName)
			fmt.Printf("Received message: %v from: %d\n", receivedMsg, ind)
		}

	}()

	// 确保接收操作能够在超时之前完成
	select {
	case <-time.After(2 * time.Second):
		t.Errorf("Timeout while waiting for the message")
	}
}

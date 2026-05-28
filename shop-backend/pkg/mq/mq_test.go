package mq

import (
	"testing"
)

func TestConnection(t *testing.T) {
	// 测试连接创建
	conn, err := NewConnection()
	if err != nil {
		t.Skipf("跳过测试: 创建连接失败: %v", err)
		return
	}
	defer conn.Close()

	// 测试通道获取
	if conn.Channel() == nil {
		t.Error("获取通道失败")
	}
}

func TestProducer(t *testing.T) {
	// 测试生产者
	conn, err := NewConnection()
	if err != nil {
		t.Skipf("跳过测试: 创建连接失败: %v", err)
		return
	}
	defer conn.Close()

	producer := NewProducer(conn)
	msg := map[string]string{"test": "message"}

	err = producer.Publish("", "test_queue", msg)
	if err != nil {
		t.Errorf("发布消息失败: %v", err)
	}
}
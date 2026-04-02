package kafka

import (
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/logc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	dialer := &kafka.Dialer{
		Timeout:   30 * time.Second,
		DualStack: true,
	}
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      brokers,
		Balancer:     &kafka.LeastBytes{},
		Async:        true, // 同步发送，方便捕获错误
		Dialer:       dialer,
		WriteTimeout: 30 * time.Second,
	})
	if topic != "" {
		w.Topic = topic
	}
	// fmt.Println("创建 Kafka 生产者，主题:", topic)
	// fmt.Println("Kafka 生产者连接的 Brokers:", brokers)
	return &KafkaProducer{writer: w}
}

func (p *KafkaProducer) Write(message *Message) error {
	// 1. 注入当前 Span 到 carrier
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(message.Ctx, carrier)
	for k, v := range carrier {
		fmt.Printf("Injected header: %s=%s\n", k, v)
	}
	// 2. 构造 Kafka 消息
	msg := kafka.Message{
		Key:   []byte(message.Key),
		Value: []byte(message.Val),
		Headers: []kafka.Header{
			{Key: "trace", Value: []byte(carrier["traceparent"])},
		},
	}
	if message.Topic != "" {
		msg.Topic = message.Topic
	}

	logc.Infof(message.Ctx, "Producing message to topic %s: key=%s, val=%s", msg.Topic, message.Key, message.Val)
	fmt.Println("Producing message to topic:", msg.Topic, "key:", message.Key, "val:", message.Val)
	// 3. 发送消息
	// ctx := context.Background()
	if err := p.writer.WriteMessages(message.Ctx, msg); err != nil {
		fmt.Println("生产消息错误:" + err.Error())
		logc.Errorf(message.Ctx, "Producing message to topic %s: key=%s, val=%s, err=%v", msg.Topic, message.Key, message.Val, err)
		return err
	}
	return nil
}

func (p *KafkaProducer) Close() {
	p.writer.Close()
}

package kafka

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/logc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// Handler 消息处理函数，支持 context
type Handler func(ctx context.Context, msg kafka.Message) error

type KafkaConsumer struct {
	Ctx       context.Context
	Reader    *kafka.Reader
	Topic     string
	Retry     int
	Handler   Handler
	DLQ       *KafkaProducer
	Buffer    chan kafka.Message
	Timeout   time.Duration
	WorkerNum int
}

// ConsumerConfig 可选配置
type ConsumerConfig struct {
	Brokers   []string
	Topic     string
	GroupID   string
	Handler   Handler
	Retry     int
	DLQ       bool
	Timeout   time.Duration
	WorkerNum int
}

func NewKafkaConsumer(ctx context.Context, cfg *ConsumerConfig) *KafkaConsumer {
	if cfg == nil {
		log.Println("ConsumerConfig is nil")
		return nil
	}
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		GroupID:  cfg.GroupID,
		Topic:    cfg.Topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	c := &KafkaConsumer{
		Ctx:       ctx,
		Topic:     cfg.Topic,
		Handler:   cfg.Handler,
		Reader:    r,
		Buffer:    make(chan kafka.Message, 1000),
		Timeout:   30 * time.Minute,
		Retry:     3,
		WorkerNum: 1,
	}

	if cfg.Retry != 0 {
		c.Retry = cfg.Retry
	}
	if cfg.DLQ {
		c.DLQ = NewKafkaProducer(cfg.Brokers, "DLQ-"+cfg.Topic)
	}
	if cfg.Timeout != 0 {
		c.Timeout = cfg.Timeout
	}
	if cfg.WorkerNum != 0 {
		c.WorkerNum = cfg.WorkerNum
	}

	return c
}

// Start 启动 Worker Pool 消费
func (c *KafkaConsumer) Start() {
	// 启动固定 WorkerPool
	for i := 0; i < c.WorkerNum; i++ {
		go c.worker(i)
	}

	// 主循环读取 Kafka 消息
	for {
		select {
		case <-c.Ctx.Done():
			logc.Infof(c.Ctx, "KafkaConsumer for topic %s exiting main loop", c.Topic)
			c.Close()
			return
		default:
			msg, err := c.Reader.ReadMessage(c.Ctx)
			if err != nil {
				log.Println("Kafka read failed:", err)
				time.Sleep(time.Second)
				continue
			}
			c.Buffer <- msg
		}
	}
}

func getHeader(msg kafka.Message, key string) string {
	for _, h := range msg.Headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

// worker 消费 Worker
func (c *KafkaConsumer) worker(id int) {
	// logc.Infof(c.Ctx, "Worker %d started", id)
	for {
		select {
		case <-c.Ctx.Done():
			logc.Infof(c.Ctx, "Worker %d exiting", id)
			return
		case message := <-c.Buffer:
			// recover 防止 panic
			func(msg kafka.Message) {
				defer func() {
					if r := recover(); r != nil {
						logc.Infof(c.Ctx, "Worker %d recovered from panic: %v, message key: %s", id, r, string(msg.Key))
					}
				}()
				logc.Infof(c.Ctx, "Worker %d message info, message key: %s", id, string(msg.Key))

				// 1. 把消息头还原成 carrier
				carrier := propagation.MapCarrier{"traceparent": string(getHeader(msg, "trace"))}
				ctx := otel.GetTextMapPropagator().Extract(c.Ctx, carrier)
				// 2. 起新的 consumer span
				tracer := otel.Tracer("kafka-consumer")
				ctx, span := tracer.Start(ctx, "kafka-consume")
				defer span.End()

				// 2. 处理消息，带重试机制
				err := c.processWithRetry(ctx, msg)
				if err == nil {
					return
				}
				// 3. 处理失败，记录日志，发送到 DLQ（如果配置了的话）

				if c.DLQ != nil {
					go func(msg kafka.Message) {
						if err := c.DLQ.Write(&Message{
							Ctx:   context.Background(),
							Key:   string(msg.Key),
							Val:   string(msg.Value),
							Topic: "DLQ-" + c.Topic,
						}); err != nil {
							logc.Infof(c.Ctx, "Failed to write DLQ: %v", err)
						}
					}(message)
				}
			}(message)
			return
		}
	}
}

// processWithRetry 消费消息 + 重试 + 指数退避
func (c *KafkaConsumer) processWithRetry(parentCtx context.Context, msg kafka.Message) error {
	// 1. 如果外部已经取消，直接退出
	if err := parentCtx.Err(); err != nil {
		return err
	}

	const baseDelay = 500 * time.Millisecond
	maxBackoff := 30 * time.Second

	var lastErr error
	for i := 0; i <= c.Retry; i++ {
		func() {
			ctx, cancel := context.WithTimeout(parentCtx, c.Timeout)
			defer cancel()
			lastErr = c.Handler(ctx, msg)
		}()

		if lastErr == nil {
			return nil
		}

		backoff := baseDelay * (1 << uint(i))
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
		select {
		case <-parentCtx.Done():
			return parentCtx.Err()
		case <-time.After(backoff):
		}
	}
	return lastErr
}

// Close 关闭消费者
func (c *KafkaConsumer) Close() {
	if c.Reader != nil {
		c.Reader.Close()
	}
	if c.DLQ != nil {
		c.DLQ.Close()
	}
}

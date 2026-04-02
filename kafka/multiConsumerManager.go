package kafka

import (
	"context"
	"fmt"
	"sync"

	"github.com/zeromicro/go-zero/core/logc"
)

type MultiConsumerManager struct {
	brokers    []string
	consumers  map[string]*KafkaConsumer
	ctxCancels map[string]context.CancelFunc
	wgs        map[string]*sync.WaitGroup
}

// NewMultiConsumerManager 创建管理器
func NewMultiConsumerManager(brokers []string) *MultiConsumerManager {
	return &MultiConsumerManager{
		brokers:    brokers,
		consumers:  make(map[string]*KafkaConsumer),
		ctxCancels: make(map[string]context.CancelFunc),
		wgs:        make(map[string]*sync.WaitGroup),
	}
}

// StartQueues 启动多个队列
func (m *MultiConsumerManager) StartQueues(queues []*ConsumerConfig) {
	for _, q := range queues {
		// 创建单独 ctx
		ctx, cancel := context.WithCancel(context.Background())
		m.ctxCancels[q.Topic] = cancel

		// 创建 WaitGroup
		wg := &sync.WaitGroup{}
		m.wgs[q.Topic] = wg

		// 创建 Consumer
		if len(q.Brokers) == 0 {
			q.Brokers = m.brokers
		}
		consumer := NewKafkaConsumer(ctx, q)
		m.consumers[q.Topic] = consumer

		logc.Infof(ctx, "[Manager] Queue registered: %s", q.Topic)

		// 启动 Consumer
		wg.Add(1)
		go func(c *KafkaConsumer, wg *sync.WaitGroup, topic string) {
			defer wg.Done()
			c.Start()
			logc.Infof(ctx, "Consumer for topic %s exited", topic)
		}(consumer, wg, q.Topic)

		logc.Infof(ctx, "Started consumer for topic: %s", q.Topic)
	}
}

// AddQueue 仅添加，不启动
func (m *MultiConsumerManager) AddQueue(q *ConsumerConfig) {
	// 创建单独 ctx
	ctx, cancel := context.WithCancel(context.Background())
	m.ctxCancels[q.Topic] = cancel

	// 创建 WaitGroup
	wg := &sync.WaitGroup{}
	m.wgs[q.Topic] = wg

	// 创建 Consumer
	if q.Brokers == nil || len(q.Brokers) == 0 {
		q.Brokers = m.brokers
	}
	consumer := NewKafkaConsumer(ctx, q)
	m.consumers[q.Topic] = consumer

	logc.Infof(ctx, "[Manager] Queue registered: %s", q.Topic)
}

// StartAll 启动所有已注册队列
func (m *MultiConsumerManager) StartAll() {
	for topic, consumer := range m.consumers {
		wg := m.wgs[topic]
		wg.Add(1)
		go func(topic string, c *KafkaConsumer, wg *sync.WaitGroup) {
			defer wg.Done()
			logc.Infof(c.Ctx, "[Manager] Starting consumer for topic: %s", topic)

			c.Start()

			logc.Infof(c.Ctx, "[Manager] Consumer exited: %s", topic)
		}(topic, consumer, wg)
	}
}

// StopQueue 优雅停止单个队列
func (m *MultiConsumerManager) StopQueue(topic string) {
	if cancel, ok := m.ctxCancels[topic]; ok {
		cancel()
		if wg, ok := m.wgs[topic]; ok {
			wg.Wait() // 等待所有 Worker 完成
		}
		delete(m.ctxCancels, topic)
		delete(m.consumers, topic)
		delete(m.wgs, topic)
		logc.Infof(context.Background(), "Stopped consumer for topic: %s", topic)
	}
}

// StopAll 优雅停止所有队列
func (m *MultiConsumerManager) StopAll() {
	fmt.Println("Stopping all consumers...")
	ctx := context.Background()
	for topic, cancel := range m.ctxCancels {
		cancel()
		logc.Infof(ctx, "Stopping consumer context for topic %s exited", topic)
		fmt.Println("Stopping consumer context for topic:", topic)
	}
	for topic, wg := range m.wgs {
		wg.Wait()
		logc.Infof(ctx, "Stopping consumer waitgroup for topic %s exited", topic)
		fmt.Println("Stopping consumer waitgroup for topic exited:", topic)
	}
	m.ctxCancels = make(map[string]context.CancelFunc)
	m.consumers = make(map[string]*KafkaConsumer)
	m.wgs = make(map[string]*sync.WaitGroup)
}

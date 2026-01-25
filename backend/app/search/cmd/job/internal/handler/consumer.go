package handler

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"polaris-io/backend/app/search/cmd/job/internal/svc"

	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/logx"
)

// KafkaConsumer Kafka 消费者
type KafkaConsumer struct {
	reader       *kafka.Reader
	handler      *FileEventHandler
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewKafkaConsumer 创建 Kafka 消费者
func NewKafkaConsumer(svcCtx *svc.ServiceContext) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  svcCtx.Config.Kafka.Brokers,
		Topic:    svcCtx.Config.Kafka.Topic,
		GroupID:  svcCtx.Config.Kafka.GroupID,
		MinBytes: 1,                // 最小拉取字节
		MaxBytes: 10 * 1024 * 1024, // 最大 10MB
	})

	ctx, cancel := context.WithCancel(context.Background())

	return &KafkaConsumer{
		reader:  reader,
		handler: NewFileEventHandler(svcCtx),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start 启动消费者
func (c *KafkaConsumer) Start() {
	logx.Info("Starting Kafka consumer...")

	// 监听系统信号
	go c.handleSignals()

	for {
		select {
		case <-c.ctx.Done():
			logx.Info("Kafka consumer stopped")
			return
		default:
			msg, err := c.reader.FetchMessage(c.ctx)
			if err != nil {
				if c.ctx.Err() != nil {
					// 上下文已取消，正常退出
					return
				}
				logx.Errorf("Failed to fetch message: %v", err)
				continue
			}

			// 处理消息
			if err := c.handler.Handle(c.ctx, msg.Key, msg.Value); err != nil {
				logx.Errorf("Failed to handle message: %v", err)
				// 处理失败也提交 offset，避免卡住
				// 生产环境可以考虑死信队列
			}

			// 提交 offset
			if err := c.reader.CommitMessages(c.ctx, msg); err != nil {
				logx.Errorf("Failed to commit message: %v", err)
			}
		}
	}
}

// handleSignals 处理系统信号
func (c *KafkaConsumer) handleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logx.Info("Received shutdown signal")
	c.Stop()
}

// Stop 停止消费者
func (c *KafkaConsumer) Stop() {
	c.cancel()
	if err := c.reader.Close(); err != nil {
		logx.Errorf("Failed to close Kafka reader: %v", err)
	}
}

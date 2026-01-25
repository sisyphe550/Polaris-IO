package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/logx"
)

// ProducerConfig Kafka 生产者配置
type ProducerConfig struct {
	Brokers []string
	Topic   string
}

// Producer Kafka 生产者
type Producer struct {
	writer *kafka.Writer
	topic  string
}

// FileEvent 文件事件
type FileEvent struct {
	EventType string `json:"event_type"` // file_uploaded, file_deleted, file_updated
	UserId    int64  `json:"user_id"`
	FileId    int64  `json:"file_id"`
	Identity  string `json:"identity"`
	Name      string `json:"name"`
	Hash      string `json:"hash,omitempty"`
	Size      uint64 `json:"size,omitempty"`
	Ext       string `json:"ext,omitempty"`
	ParentId  int64  `json:"parent_id"` // 不使用 omitempty，0 表示根目录
	Timestamp string `json:"timestamp"`
}

// NewProducer 创建 Kafka 生产者
func NewProducer(cfg ProducerConfig) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
		Async:        true, // 异步发送，提高性能
	}

	return &Producer{
		writer: writer,
		topic:  cfg.Topic,
	}
}

// SendFileEvent 发送文件事件
func (p *Producer) SendFileEvent(ctx context.Context, event *FileEvent) error {
	if event.Timestamp == "" {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(event.Identity), // 使用 identity 作为 key，确保同一文件的事件顺序
		Value: data,
	}

	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		logx.Errorf("Failed to send file event to kafka: %v", err)
		return err
	}

	logx.Infof("Sent file event: type=%s, identity=%s", event.EventType, event.Identity)
	return nil
}

// SendFileUploaded 发送文件上传完成事件
func (p *Producer) SendFileUploaded(ctx context.Context, userId, fileId int64, identity, name, hash string, size uint64, ext string) error {
	return p.SendFileEvent(ctx, &FileEvent{
		EventType: "file_uploaded",
		UserId:    userId,
		FileId:    fileId,
		Identity:  identity,
		Name:      name,
		Hash:      hash,
		Size:      size,
		Ext:       ext,
	})
}

// SendFileDeleted 发送文件删除事件
func (p *Producer) SendFileDeleted(ctx context.Context, userId, fileId int64, identity string) error {
	return p.SendFileEvent(ctx, &FileEvent{
		EventType: "file_deleted",
		UserId:    userId,
		FileId:    fileId,
		Identity:  identity,
	})
}

// SendFileUpdated 发送文件更新事件 (移动/重命名)
func (p *Producer) SendFileUpdated(ctx context.Context, userId, fileId int64, identity, name string, parentId int64) error {
	return p.SendFileEvent(ctx, &FileEvent{
		EventType: "file_updated",
		UserId:    userId,
		FileId:    fileId,
		Identity:  identity,
		Name:      name,
		ParentId:  parentId,
	})
}

// Close 关闭生产者
func (p *Producer) Close() error {
	return p.writer.Close()
}

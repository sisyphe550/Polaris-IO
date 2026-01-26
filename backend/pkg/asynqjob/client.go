package asynqjob

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/logx"
)

// AsynqClientConfig asynq 客户端配置
type AsynqClientConfig struct {
	Addr     string // Redis 地址
	Password string // Redis 密码
}

// AsynqClient asynq 客户端封装
type AsynqClient struct {
	client *asynq.Client
}

// NewAsynqClient 创建 asynq 客户端
func NewAsynqClient(cfg AsynqClientConfig) *AsynqClient {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Addr,
		Password: cfg.Password,
	})
	return &AsynqClient{client: client}
}

// Close 关闭客户端
func (c *AsynqClient) Close() error {
	return c.client.Close()
}

// EnqueueS3Cleanup 入队 S3 清理任务
func (c *AsynqClient) EnqueueS3Cleanup(ctx context.Context, payload S3CleanupPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(TypeS3Cleanup, data)

	// 立即执行，最多重试 3 次
	info, err := c.client.EnqueueContext(ctx, task,
		asynq.MaxRetry(3),
		asynq.Queue("critical"),
	)
	if err != nil {
		logx.WithContext(ctx).Errorf("EnqueueS3Cleanup failed: %v", err)
		return err
	}

	logx.WithContext(ctx).Infof("EnqueueS3Cleanup success: taskId=%s, hash=%s", info.ID, payload.Hash)
	return nil
}

// EnqueueTrashClear 入队回收站清理任务
func (c *AsynqClient) EnqueueTrashClear(ctx context.Context, payload TrashClearPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(TypeTrashClear, data)

	// 立即执行，最多重试 5 次
	info, err := c.client.EnqueueContext(ctx, task,
		asynq.MaxRetry(5),
		asynq.Queue("default"),
	)
	if err != nil {
		logx.WithContext(ctx).Errorf("EnqueueTrashClear failed: %v", err)
		return err
	}

	logx.WithContext(ctx).Infof("EnqueueTrashClear success: taskId=%s, userId=%d", info.ID, payload.UserId)
	return nil
}

// EnqueueUploadTimeout 入队上传超时检测任务（延迟执行）
func (c *AsynqClient) EnqueueUploadTimeout(ctx context.Context, payload UploadTimeoutPayload, delay time.Duration) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(TypeUploadTimeout, data)

	// 延迟执行，最多重试 1 次
	info, err := c.client.EnqueueContext(ctx, task,
		asynq.ProcessIn(delay),
		asynq.MaxRetry(1),
		asynq.Queue("low"),
	)
	if err != nil {
		logx.WithContext(ctx).Errorf("EnqueueUploadTimeout failed: %v", err)
		return err
	}

	logx.WithContext(ctx).Infof("EnqueueUploadTimeout success: taskId=%s, uploadKey=%s, delay=%v",
		info.ID, payload.UploadKey, delay)
	return nil
}

// EnqueueQuotaRefund 入队配额退还任务
func (c *AsynqClient) EnqueueQuotaRefund(ctx context.Context, payload QuotaRefundPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(TypeQuotaRefund, data)

	// 立即执行，最多重试 5 次
	info, err := c.client.EnqueueContext(ctx, task,
		asynq.MaxRetry(5),
		asynq.Queue("critical"),
	)
	if err != nil {
		logx.WithContext(ctx).Errorf("EnqueueQuotaRefund failed: %v", err)
		return err
	}

	logx.WithContext(ctx).Infof("EnqueueQuotaRefund success: taskId=%s, userId=%d, size=%d",
		info.ID, payload.UserId, payload.Size)
	return nil
}

// EnqueueShareExpire 入队分享过期任务（延迟执行）
func (c *AsynqClient) EnqueueShareExpire(ctx context.Context, payload ShareExpirePayload, delay time.Duration) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(TypeShareExpire, data)

	// 延迟执行，最多重试 1 次
	info, err := c.client.EnqueueContext(ctx, task,
		asynq.ProcessIn(delay),
		asynq.MaxRetry(1),
		asynq.Queue("low"),
	)
	if err != nil {
		logx.WithContext(ctx).Errorf("EnqueueShareExpire failed: %v", err)
		return err
	}

	logx.WithContext(ctx).Infof("EnqueueShareExpire success: taskId=%s, shareIdentity=%s, delay=%v",
		info.ID, payload.ShareIdentity, delay)
	return nil
}

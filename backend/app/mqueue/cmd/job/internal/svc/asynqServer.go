package svc

import (
	"context"

	"polaris-io/backend/app/mqueue/cmd/job/internal/config"

	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/logx"
)

// NewAsynqServer 创建 asynq server
func NewAsynqServer(c config.Config) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     c.Redis.Addr,
			Password: c.Redis.Password,
		},
		asynq.Config{
			// 最大并发处理任务数
			Concurrency: 10,
			// 队列优先级配置
			Queues: map[string]int{
				"critical": 6, // 高优先级（S3 清理）
				"default":  3, // 默认优先级（回收站清理）
				"low":      1, // 低优先级（上传超时检测/延迟任务）
			},
			// 错误处理
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				logx.Errorf("asynq task failed: type=%s, err=%v", task.Type(), err)
			}),
		},
	)
}

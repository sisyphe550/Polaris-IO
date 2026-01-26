package handler

import (
	"polaris-io/backend/app/mqueue/cmd/job/internal/svc"
	"polaris-io/backend/pkg/asynqjob"

	"github.com/hibiken/asynq"
)

// RegisterHandlers 注册所有任务处理器
func RegisterHandlers(mux *asynq.ServeMux, svcCtx *svc.ServiceContext, asynqClient *asynq.Client) {
	// S3 文件清理任务（高优先级）
	mux.HandleFunc(asynqjob.TypeS3Cleanup, NewS3CleanupHandler(svcCtx).ProcessTask)

	// 回收站批量清理任务（默认优先级）
	mux.HandleFunc(asynqjob.TypeTrashClear, NewTrashClearHandler(svcCtx, asynqClient).ProcessTask)

	// 上传超时检测任务（低优先级）
	mux.HandleFunc(asynqjob.TypeUploadTimeout, NewUploadTimeoutHandler(svcCtx, asynqClient).ProcessTask)

	// 配额退还任务（默认优先级）
	mux.HandleFunc(asynqjob.TypeQuotaRefund, NewQuotaRefundHandler(svcCtx).ProcessTask)

	// 分享过期任务（低优先级）
	mux.HandleFunc(asynqjob.TypeShareExpire, NewShareExpireHandler(svcCtx).ProcessTask)
}

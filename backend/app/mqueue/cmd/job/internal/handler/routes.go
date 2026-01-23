package handler

import (
	"polaris-io/backend/app/mqueue/cmd/job/internal/svc"
	"polaris-io/backend/pkg/asynqjob"

	"github.com/hibiken/asynq"
)

// RegisterHandlers 注册所有任务处理器
func RegisterHandlers(mux *asynq.ServeMux, svcCtx *svc.ServiceContext, asynqClient *asynq.Client) {
	// S3 文件清理任务
	mux.HandleFunc(asynqjob.TypeS3Cleanup, NewS3CleanupHandler(svcCtx).ProcessTask)

	// 回收站批量清理任务
	mux.HandleFunc(asynqjob.TypeTrashClear, NewTrashClearHandler(svcCtx, asynqClient).ProcessTask)

	// TODO: 上传超时检测任务
	// mux.HandleFunc(asynqjob.TypeUploadTimeout, NewUploadTimeoutHandler(svcCtx).ProcessTask)

	// TODO: 配额退还任务
	// mux.HandleFunc(asynqjob.TypeQuotaRefund, NewQuotaRefundHandler(svcCtx).ProcessTask)
}

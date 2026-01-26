package handler

import (
	"context"
	"encoding/json"

	"polaris-io/backend/app/mqueue/cmd/job/internal/svc"
	"polaris-io/backend/pkg/asynqjob"

	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/logx"
)

// UploadTimeoutHandler 上传超时检测处理器
type UploadTimeoutHandler struct {
	svcCtx      *svc.ServiceContext
	asynqClient *asynq.Client
}

// NewUploadTimeoutHandler 创建上传超时检测处理器
func NewUploadTimeoutHandler(svcCtx *svc.ServiceContext, asynqClient *asynq.Client) *UploadTimeoutHandler {
	return &UploadTimeoutHandler{
		svcCtx:      svcCtx,
		asynqClient: asynqClient,
	}
}

// ProcessTask 处理上传超时检测任务
// 该任务在用户请求上传凭证时延迟创建（通常 24 小时）
// 如果在延迟时间后文件仍未上传完成，则退还预扣的配额
func (h *UploadTimeoutHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload asynqjob.UploadTimeoutPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		logx.Errorf("UploadTimeoutHandler: failed to unmarshal payload: %v", err)
		return err
	}

	logx.Infof("UploadTimeoutHandler: processing task, userId=%d, hash=%s, size=%d",
		payload.UserId, payload.Hash, payload.Size)

	// 参数校验
	if payload.UserId <= 0 || payload.Hash == "" || payload.Size == 0 {
		logx.Errorf("UploadTimeoutHandler: invalid payload, userId=%d, hash=%s, size=%d",
			payload.UserId, payload.Hash, payload.Size)
		return nil // 无效参数，不重试
	}

	// 检查 file_meta 是否存在该 hash（上传是否完成）
	meta, err := h.svcCtx.FileMetaModel.FindByHash(ctx, payload.Hash)
	if err != nil {
		// 查询出错，记录日志，返回错误让 asynq 重试
		logx.Errorf("UploadTimeoutHandler: FindByHash error: %v, hash=%s", err, payload.Hash)
		return err
	}

	if meta != nil {
		// 文件已上传完成，无需退还配额
		logx.Infof("UploadTimeoutHandler: file upload completed, skip refund, userId=%d, hash=%s",
			payload.UserId, payload.Hash)
		return nil
	}

	// 文件未上传完成，需要退还配额
	logx.Infof("UploadTimeoutHandler: file upload not completed, refunding quota, userId=%d, hash=%s, size=%d",
		payload.UserId, payload.Hash, payload.Size)

	// 入队配额退还任务
	refundPayload := asynqjob.QuotaRefundPayload{
		UserId: payload.UserId,
		Size:   payload.Size,
		Reason: "upload_timeout",
	}

	data, err := json.Marshal(refundPayload)
	if err != nil {
		logx.Errorf("UploadTimeoutHandler: failed to marshal QuotaRefundPayload: %v", err)
		return err
	}

	refundTask := asynq.NewTask(asynqjob.TypeQuotaRefund, data)
	_, err = h.asynqClient.EnqueueContext(ctx, refundTask,
		asynq.MaxRetry(3),
		asynq.Queue("default"),
	)
	if err != nil {
		logx.Errorf("UploadTimeoutHandler: failed to enqueue QuotaRefund task: %v", err)
		return err
	}

	logx.Infof("UploadTimeoutHandler: QuotaRefund task enqueued, userId=%d, size=%d",
		payload.UserId, payload.Size)

	return nil
}

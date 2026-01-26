package handler

import (
	"context"
	"encoding/json"

	"polaris-io/backend/app/mqueue/cmd/job/internal/svc"
	"polaris-io/backend/app/user/cmd/rpc/usercenter"
	"polaris-io/backend/pkg/asynqjob"

	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/logx"
)

// QuotaRefundHandler 配额退还处理器
type QuotaRefundHandler struct {
	svcCtx *svc.ServiceContext
}

// NewQuotaRefundHandler 创建配额退还处理器
func NewQuotaRefundHandler(svcCtx *svc.ServiceContext) *QuotaRefundHandler {
	return &QuotaRefundHandler{svcCtx: svcCtx}
}

// ProcessTask 处理配额退还任务
func (h *QuotaRefundHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload asynqjob.QuotaRefundPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		logx.Errorf("QuotaRefundHandler: failed to unmarshal payload: %v", err)
		return err
	}

	logx.Infof("QuotaRefundHandler: processing task, userId=%d, size=%d, reason=%s",
		payload.UserId, payload.Size, payload.Reason)

	// 参数校验
	if payload.UserId <= 0 || payload.Size == 0 {
		logx.Errorf("QuotaRefundHandler: invalid payload, userId=%d, size=%d",
			payload.UserId, payload.Size)
		return nil // 无效参数，不重试
	}

	// 调用 UsercenterRpc 退还配额
	_, err := h.svcCtx.UsercenterRpc.RefundQuota(ctx, &usercenter.RefundQuotaReq{
		UserId: payload.UserId,
		Size:   payload.Size,
	})
	if err != nil {
		logx.Errorf("QuotaRefundHandler: RefundQuota failed: %v, userId=%d, size=%d",
			err, payload.UserId, payload.Size)
		return err // 返回错误，asynq 会重试
	}

	logx.Infof("QuotaRefundHandler: quota refunded successfully, userId=%d, size=%d, reason=%s",
		payload.UserId, payload.Size, payload.Reason)

	return nil
}

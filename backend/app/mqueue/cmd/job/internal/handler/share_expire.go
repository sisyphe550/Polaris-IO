package handler

import (
	"context"
	"encoding/json"
	"errors"

	shareModel "polaris-io/backend/app/share/model"

	"polaris-io/backend/app/mqueue/cmd/job/internal/svc"
	"polaris-io/backend/pkg/asynqjob"

	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/logx"
)

// ShareExpireHandler 分享过期处理器
type ShareExpireHandler struct {
	svcCtx *svc.ServiceContext
}

// NewShareExpireHandler 创建分享过期处理器
func NewShareExpireHandler(svcCtx *svc.ServiceContext) *ShareExpireHandler {
	return &ShareExpireHandler{svcCtx: svcCtx}
}

// ProcessTask 处理分享过期任务
// 该任务在创建分享时根据过期时间延迟创建
// 到期后将分享状态置为已过期（软删除）
func (h *ShareExpireHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload asynqjob.ShareExpirePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		logx.Errorf("ShareExpireHandler: failed to unmarshal payload: %v", err)
		return err
	}

	logx.Infof("ShareExpireHandler: processing task, shareIdentity=%s, userId=%d",
		payload.ShareIdentity, payload.UserId)

	// 参数校验
	if payload.ShareIdentity == "" {
		logx.Errorf("ShareExpireHandler: invalid payload, shareIdentity is empty")
		return nil // 无效参数，不重试
	}

	// 将分享设置为过期状态
	err := h.svcCtx.ShareModel.SetExpiredByIdentity(ctx, payload.ShareIdentity)
	if err != nil {
		if errors.Is(err, shareModel.ErrNotFound) {
			// 分享不存在或已经被删除/过期，无需处理
			logx.Infof("ShareExpireHandler: share not found or already expired, shareIdentity=%s",
				payload.ShareIdentity)
			return nil
		}
		logx.Errorf("ShareExpireHandler: SetExpiredByIdentity failed: %v, shareIdentity=%s",
			err, payload.ShareIdentity)
		return err // 返回错误，asynq 会重试
	}

	logx.Infof("ShareExpireHandler: share expired successfully, shareIdentity=%s, userId=%d",
		payload.ShareIdentity, payload.UserId)

	return nil
}

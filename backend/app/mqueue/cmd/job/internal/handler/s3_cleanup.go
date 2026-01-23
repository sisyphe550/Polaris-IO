package handler

import (
	"context"
	"encoding/json"

	"polaris-io/backend/app/mqueue/cmd/job/internal/svc"
	"polaris-io/backend/pkg/asynqjob"

	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/logx"
)

// S3CleanupHandler S3 文件清理处理器
type S3CleanupHandler struct {
	svcCtx *svc.ServiceContext
}

// NewS3CleanupHandler 创建 S3 清理处理器
func NewS3CleanupHandler(svcCtx *svc.ServiceContext) *S3CleanupHandler {
	return &S3CleanupHandler{svcCtx: svcCtx}
}

// ProcessTask 处理 S3 清理任务
func (h *S3CleanupHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload asynqjob.S3CleanupPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		logx.Errorf("S3CleanupHandler: failed to unmarshal payload: %v", err)
		return err
	}

	logx.Infof("S3CleanupHandler: processing task, hash=%s, s3Key=%s", payload.Hash, payload.S3Key)

	// 1. 再次检查 file_meta 的引用计数（防止并发问题）
	meta, err := h.svcCtx.FileMetaModel.FindByHash(ctx, payload.Hash)
	if err != nil {
		// 如果找不到记录，说明已经被删除了，跳过
		logx.Infof("S3CleanupHandler: file_meta not found, maybe already deleted, hash=%s", payload.Hash)
		return nil
	}

	// 如果引用计数 > 0，说明又有新的引用了，不删除
	if meta.RefCount > 0 {
		logx.Infof("S3CleanupHandler: ref_count > 0, skip cleanup, hash=%s, ref_count=%d",
			payload.Hash, meta.RefCount)
		return nil
	}

	// 2. 删除 S3 文件
	if err := h.svcCtx.S3Client.DeleteObject(ctx, payload.S3Key); err != nil {
		logx.Errorf("S3CleanupHandler: failed to delete S3 object: %v, s3Key=%s", err, payload.S3Key)
		return err // 返回错误，asynq 会重试
	}

	logx.Infof("S3CleanupHandler: S3 object deleted successfully, s3Key=%s", payload.S3Key)

	// 3. 删除 MongoDB file_meta 记录
	if err := h.svcCtx.FileMetaModel.DeleteByHash(ctx, payload.Hash); err != nil {
		logx.Errorf("S3CleanupHandler: failed to delete file_meta: %v, hash=%s", err, payload.Hash)
		// 不返回错误，S3 文件已删除，记录删除失败只记录日志
	} else {
		logx.Infof("S3CleanupHandler: file_meta deleted successfully, hash=%s", payload.Hash)
	}

	return nil
}

package handler

import (
	"context"
	"encoding/json"
	"errors"

	fileMongo "polaris-io/backend/app/file/mongo"
	"polaris-io/backend/app/mqueue/cmd/job/internal/svc"
	"polaris-io/backend/pkg/asynqjob"

	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/logx"
)

// TrashClearHandler 回收站清理处理器
type TrashClearHandler struct {
	svcCtx      *svc.ServiceContext
	asynqClient *asynq.Client
}

// NewTrashClearHandler 创建回收站清理处理器
func NewTrashClearHandler(svcCtx *svc.ServiceContext, asynqClient *asynq.Client) *TrashClearHandler {
	return &TrashClearHandler{
		svcCtx:      svcCtx,
		asynqClient: asynqClient,
	}
}

// ProcessTask 处理回收站清理任务
func (h *TrashClearHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload asynqjob.TrashClearPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		logx.Errorf("TrashClearHandler: failed to unmarshal payload: %v", err)
		return err
	}

	logx.Infof("TrashClearHandler: processing task, userId=%d, batchSize=%d",
		payload.UserId, payload.BatchSize)

	if payload.BatchSize <= 0 {
		payload.BatchSize = 50
	}

	var totalDeleted int64

	// 分批处理
	for {
		// 查询一批待删除的文件
		files, _, err := h.svcCtx.UserRepositoryModel.FindTrashList(ctx, uint64(payload.UserId), 1, int64(payload.BatchSize))
		if err != nil {
			logx.Errorf("TrashClearHandler: FindTrashList error: %v", err)
			return err
		}

		if len(files) == 0 {
			break
		}

		for _, file := range files {
			// 处理文件（减少引用计数，发送 S3 清理任务）
			if file.Hash != "" {
				meta, err := h.svcCtx.FileMetaModel.DecrRefCountAndGet(ctx, file.Hash, 1)
				if err != nil {
					// 根据错误类型分别处理
					switch {
					case errors.Is(err, fileMongo.ErrNotFound):
						// MongoDB 中没有这条记录，数据不一致，但不影响删除
						logx.Infof("TrashClearHandler: file_meta not found in MongoDB, hash=%s", file.Hash)
					case errors.Is(err, fileMongo.ErrRefCountZero):
						// 引用计数已经为 0，说明已经被扣减过了，检查是否需要清理 S3
						logx.Infof("TrashClearHandler: ref_count already zero, hash=%s", file.Hash)
						if meta != nil && meta.RefCount <= 0 {
							h.enqueueS3Cleanup(ctx, asynqjob.S3CleanupPayload{
								Hash:   file.Hash,
								S3Key:  meta.S3Key,
								UserId: payload.UserId,
							})
						}
					default:
						logx.Errorf("TrashClearHandler: DecrRefCountAndGet error: %v, hash=%s", err, file.Hash)
					}
				} else if meta.RefCount <= 0 {
					// 发送 S3 清理任务
					h.enqueueS3Cleanup(ctx, asynqjob.S3CleanupPayload{
						Hash:   file.Hash,
						S3Key:  meta.S3Key,
						UserId: payload.UserId,
					})
				}
			}

			// 彻底删除 MySQL 记录
			if err := h.svcCtx.UserRepositoryModel.HardDelete(ctx, nil, file.Id); err != nil {
				logx.Errorf("TrashClearHandler: HardDelete error: %v, fileId=%d", err, file.Id)
				continue
			}

			totalDeleted++
		}

		logx.Infof("TrashClearHandler: batch completed, userId=%d, deleted=%d, total=%d",
			payload.UserId, len(files), totalDeleted)
	}

	logx.Infof("TrashClearHandler: task completed, userId=%d, totalDeleted=%d",
		payload.UserId, totalDeleted)

	return nil
}

// enqueueS3Cleanup 发送 S3 清理任务
func (h *TrashClearHandler) enqueueS3Cleanup(ctx context.Context, payload asynqjob.S3CleanupPayload) {
	data, err := json.Marshal(payload)
	if err != nil {
		logx.Errorf("TrashClearHandler: failed to marshal S3CleanupPayload: %v", err)
		return
	}

	task := asynq.NewTask(asynqjob.TypeS3Cleanup, data)
	_, err = h.asynqClient.EnqueueContext(ctx, task,
		asynq.MaxRetry(3),
		asynq.Queue("critical"),
	)
	if err != nil {
		logx.Errorf("TrashClearHandler: failed to enqueue S3Cleanup task: %v", err)
	}
}

package logic

import (
	"context"
	"errors"

	"polaris-io/backend/app/file/cmd/rpc/internal/svc"
	"polaris-io/backend/app/file/cmd/rpc/pb"
	fileMongo "polaris-io/backend/app/file/mongo"
	"polaris-io/backend/pkg/asynqjob"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// 异步处理阈值：超过此数量的文件使用异步任务处理
	asyncThreshold = 100
	// 同步处理批次大小
	syncBatchSize = 50
)

type ClearTrashLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewClearTrashLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearTrashLogic {
	return &ClearTrashLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ClearTrash 清空回收站
// 如果回收站文件数量超过阈值，使用异步任务分批处理
func (l *ClearTrashLogic) ClearTrash(in *pb.ClearTrashReq) (*pb.ClearTrashResp, error) {
	if in.UserId == 0 {
		return nil, errors.New("userId is required")
	}

	// 首先查询回收站总数
	_, total, err := l.svcCtx.UserRepositoryModel.FindTrashList(l.ctx, uint64(in.UserId), 1, 1)
	if err != nil {
		l.Logger.Errorf("ClearTrash FindTrashList count error: %v", err)
		return nil, err
	}

	// 如果回收站为空，直接返回
	if total == 0 {
		return &pb.ClearTrashResp{DeletedCount: 0}, nil
	}

	// 如果文件数量超过阈值，并且 asynq 客户端可用，使用异步任务处理
	if total > asyncThreshold && l.svcCtx.AsynqClient != nil {
		l.Logger.Infof("ClearTrash: total=%d > threshold=%d, using async task", total, asyncThreshold)

		// 发送异步任务
		if err := l.svcCtx.AsynqClient.EnqueueTrashClear(l.ctx, asynqjob.TrashClearPayload{
			UserId:    in.UserId,
			BatchSize: syncBatchSize,
		}); err != nil {
			l.Logger.Errorf("ClearTrash EnqueueTrashClear error: %v", err)
			// 异步任务入队失败，降级到同步处理
		} else {
			// 异步任务入队成功，返回预估的删除数量
			return &pb.ClearTrashResp{
				DeletedCount: total,
			}, nil
		}
	}

	// 同步处理：分批查询并删除
	var deletedCount int64
	for {
		files, _, err := l.svcCtx.UserRepositoryModel.FindTrashList(l.ctx, uint64(in.UserId), 1, int64(syncBatchSize))
		if err != nil {
			l.Logger.Errorf("ClearTrash FindTrashList error: %v", err)
			return nil, err
		}

		if len(files) == 0 {
			break
		}

		for _, file := range files {
			// 减少 MongoDB file_meta 的引用计数并获取更新后的记录
			if file.Hash != "" {
				meta, err := l.svcCtx.FileMetaModel.DecrRefCountAndGet(l.ctx, file.Hash, 1)
				if err != nil {
					// 根据错误类型分别处理
					switch {
					case errors.Is(err, fileMongo.ErrNotFound):
						// MongoDB 中没有这条记录，数据不一致，但不影响删除
						l.Logger.Infof("ClearTrash: file_meta not found in MongoDB, hash=%s", file.Hash)
					case errors.Is(err, fileMongo.ErrRefCountZero):
						// 引用计数已经为 0，说明已经被扣减过了，检查是否需要清理 S3
						l.Logger.Infof("ClearTrash: ref_count already zero, hash=%s", file.Hash)
						if meta != nil && meta.RefCount <= 0 && l.svcCtx.AsynqClient != nil {
							_ = l.svcCtx.AsynqClient.EnqueueS3Cleanup(l.ctx, asynqjob.S3CleanupPayload{
								Hash:   file.Hash,
								S3Key:  meta.S3Key,
								UserId: in.UserId,
							})
						}
					default:
						l.Logger.Errorf("ClearTrash DecrRefCountAndGet error: %v, hash=%s", err, file.Hash)
					}
				} else {
					// 清除秒传缓存
					if l.svcCtx.FileCache != nil {
						_ = l.svcCtx.FileCache.DeleteFileMeta(l.ctx, file.Hash)
					}

					// 如果引用计数为 0，发送异步任务删除 S3 文件
					if meta.RefCount <= 0 && l.svcCtx.AsynqClient != nil {
						_ = l.svcCtx.AsynqClient.EnqueueS3Cleanup(l.ctx, asynqjob.S3CleanupPayload{
							Hash:   file.Hash,
							S3Key:  meta.S3Key,
							UserId: in.UserId,
						})
					}
				}
			}

			// 彻底删除 MySQL 记录
			if err := l.svcCtx.UserRepositoryModel.HardDelete(l.ctx, nil, file.Id); err != nil {
				l.Logger.Errorf("ClearTrash HardDelete error: %v", err)
				continue
			}

			deletedCount++
		}
	}

	return &pb.ClearTrashResp{
		DeletedCount: deletedCount,
	}, nil
}

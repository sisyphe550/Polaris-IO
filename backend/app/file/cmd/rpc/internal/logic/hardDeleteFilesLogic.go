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

type HardDeleteFilesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHardDeleteFilesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HardDeleteFilesLogic {
	return &HardDeleteFilesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// HardDeleteFiles 彻底删除文件
func (l *HardDeleteFilesLogic) HardDeleteFiles(in *pb.HardDeleteFilesReq) (*pb.HardDeleteFilesResp, error) {
	if in.UserId == 0 || len(in.Identities) == 0 {
		return nil, errors.New("userId and identities are required")
	}

	var deletedCount int64

	// 批量查询已删除的文件
	files, err := l.svcCtx.UserRepositoryModel.FindDeletedByIdentities(l.ctx, uint64(in.UserId), in.Identities)
	if err != nil {
		l.Logger.Errorf("HardDeleteFiles FindDeletedByIdentities error: %v", err)
		return nil, err
	}

	for _, file := range files {
		// 如果是文件夹，递归删除子文件
		if file.Ext == "" && file.Hash == "" {
			subDeletedCount := l.hardDeleteFolder(file.Id, in.UserId)
			deletedCount += subDeletedCount
		} else {
			// 减少 MongoDB file_meta 的引用计数并获取更新后的记录
			if file.Hash != "" {
				meta, err := l.svcCtx.FileMetaModel.DecrRefCountAndGet(l.ctx, file.Hash, 1)
				if err != nil {
					// 根据错误类型分别处理
					switch {
					case errors.Is(err, fileMongo.ErrNotFound):
						// MongoDB 中没有这条记录，数据不一致，但不影响删除
						l.Logger.Infof("HardDeleteFiles: file_meta not found in MongoDB, hash=%s", file.Hash)
					case errors.Is(err, fileMongo.ErrRefCountZero):
						// 引用计数已经为 0，说明已经被扣减过了，检查是否需要清理 S3
						l.Logger.Infof("HardDeleteFiles: ref_count already zero, hash=%s", file.Hash)
						if meta != nil && meta.RefCount <= 0 && l.svcCtx.AsynqClient != nil {
							l.Logger.Infof("HardDeleteFiles: scheduling S3 cleanup for hash=%s, s3Key=%s",
								file.Hash, meta.S3Key)
							_ = l.svcCtx.AsynqClient.EnqueueS3Cleanup(l.ctx, asynqjob.S3CleanupPayload{
								Hash:   file.Hash,
								S3Key:  meta.S3Key,
								UserId: in.UserId,
							})
						}
					default:
						l.Logger.Errorf("HardDeleteFiles DecrRefCountAndGet error: %v, hash=%s", err, file.Hash)
					}
					// 继续处理，不影响删除
				} else {
					// 清除秒传缓存
					if l.svcCtx.FileCache != nil {
						if err := l.svcCtx.FileCache.DeleteFileMeta(l.ctx, file.Hash); err != nil {
							l.Logger.Errorf("HardDeleteFiles DeleteFileMeta error: %v", err)
						}
					}

					// 如果引用计数为 0，发送异步任务删除 S3 文件
					if meta.RefCount <= 0 {
						l.Logger.Infof("HardDeleteFiles: ref_count=0, scheduling S3 cleanup for hash=%s, s3Key=%s",
							file.Hash, meta.S3Key)

						if l.svcCtx.AsynqClient != nil {
							if err := l.svcCtx.AsynqClient.EnqueueS3Cleanup(l.ctx, asynqjob.S3CleanupPayload{
								Hash:   file.Hash,
								S3Key:  meta.S3Key,
								UserId: in.UserId,
							}); err != nil {
								l.Logger.Errorf("HardDeleteFiles EnqueueS3Cleanup error: %v", err)
								// 任务入队失败，记录日志但不影响删除
							}
						}
					}
				}
			}
		}

		// 彻底删除 MySQL 记录
		if err := l.svcCtx.UserRepositoryModel.HardDelete(l.ctx, nil, file.Id); err != nil {
			l.Logger.Errorf("HardDeleteFiles HardDelete error: %v", err)
			continue
		}

		deletedCount++
	}

	return &pb.HardDeleteFilesResp{
		DeletedCount: deletedCount,
	}, nil
}

// hardDeleteFolder 递归彻底删除文件夹内容
func (l *HardDeleteFilesLogic) hardDeleteFolder(folderId uint64, userId int64) int64 {
	var deletedCount int64

	// 查询文件夹下已删除的子文件/文件夹
	// 使用 parent_id 查询
	files, _, err := l.svcCtx.UserRepositoryModel.FindTrashList(l.ctx, uint64(userId), 1, 1000)
	if err != nil {
		l.Logger.Errorf("hardDeleteFolder FindTrashList error: %v", err)
		return 0
	}

	for _, file := range files {
		// 只处理属于这个文件夹的文件
		if file.ParentId != folderId {
			continue
		}

		// 如果是子文件夹，递归删除
		if file.Ext == "" && file.Hash == "" {
			subDeletedCount := l.hardDeleteFolder(file.Id, userId)
			deletedCount += subDeletedCount
		} else {
			// 减少引用计数并获取更新后的记录
			if file.Hash != "" {
				meta, err := l.svcCtx.FileMetaModel.DecrRefCountAndGet(l.ctx, file.Hash, 1)
				if err != nil {
					// 根据错误类型分别处理
					switch {
					case errors.Is(err, fileMongo.ErrNotFound):
						l.Logger.Infof("hardDeleteFolder: file_meta not found in MongoDB, hash=%s", file.Hash)
					case errors.Is(err, fileMongo.ErrRefCountZero):
						l.Logger.Infof("hardDeleteFolder: ref_count already zero, hash=%s", file.Hash)
						if meta != nil && meta.RefCount <= 0 && l.svcCtx.AsynqClient != nil {
							_ = l.svcCtx.AsynqClient.EnqueueS3Cleanup(l.ctx, asynqjob.S3CleanupPayload{
								Hash:   file.Hash,
								S3Key:  meta.S3Key,
								UserId: userId,
							})
						}
					default:
						l.Logger.Errorf("hardDeleteFolder DecrRefCountAndGet error: %v, hash=%s", err, file.Hash)
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
							UserId: userId,
						})
					}
				}
			}
		}

		// 删除记录
		if err := l.svcCtx.UserRepositoryModel.HardDelete(l.ctx, nil, file.Id); err != nil {
			l.Logger.Errorf("hardDeleteFolder HardDelete error: %v", err)
			continue
		}

		deletedCount++
	}

	return deletedCount
}

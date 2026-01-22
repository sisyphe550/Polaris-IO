package logic

import (
	"context"
	"errors"

	"polaris-io/backend/app/file/cmd/rpc/internal/svc"
	"polaris-io/backend/app/file/cmd/rpc/pb"
	"polaris-io/backend/app/file/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type RestoreFilesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRestoreFilesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RestoreFilesLogic {
	return &RestoreFilesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RestoreFiles 恢复文件 (从回收站恢复)
func (l *RestoreFilesLogic) RestoreFiles(in *pb.RestoreFilesReq) (*pb.RestoreFilesResp, error) {
	if in.UserId == 0 || len(in.Identities) == 0 {
		return nil, errors.New("userId and identities are required")
	}

	var restoredCount int64
	var usedSize uint64
	affectedParentIds := make(map[int64]struct{}) // 记录受影响的父目录

	// 批量查询已删除的文件
	files, err := l.svcCtx.UserRepositoryModel.FindDeletedByIdentities(l.ctx, uint64(in.UserId), in.Identities)
	if err != nil {
		l.Logger.Errorf("RestoreFiles FindDeletedByIdentities error: %v", err)
		return nil, err
	}

	for _, file := range files {
		// 检查原父目录是否存在
		if file.ParentId > 0 {
			_, err := l.svcCtx.UserRepositoryModel.FindOne(l.ctx, file.ParentId)
			if err != nil {
				if errors.Is(err, model.ErrNotFound) {
					// 父目录不存在或已删除，恢复到根目录
					file.ParentId = 0
				}
			}
		}

		// 记录受影响的父目录
		affectedParentIds[int64(file.ParentId)] = struct{}{}

		// 恢复文件
		if err := l.svcCtx.UserRepositoryModel.RestoreFile(l.ctx, nil, file); err != nil {
			l.Logger.Errorf("RestoreFiles RestoreFile error: %v", err)
			continue
		}

		restoredCount++
		usedSize += file.Size

		// 如果是文件夹，递归恢复子文件
		if file.Ext == "" && file.Hash == "" {
			subUsedSize, subRestoredCount := l.restoreFolder(file.Id, in.UserId)
			usedSize += subUsedSize
			restoredCount += subRestoredCount
		}

		// 发送 Kafka 事件（文件恢复）
		if err := l.svcCtx.KafkaProducer.SendFileUploaded(
			l.ctx, in.UserId, int64(file.Id), file.Identity, file.Name, file.Hash, file.Size, file.Ext); err != nil {
			l.Logger.Errorf("RestoreFiles SendFileUploaded error: %v", err)
		}
	}

	// 清除文件列表缓存
	if l.svcCtx.FileCache != nil && restoredCount > 0 {
		for parentId := range affectedParentIds {
			if err := l.svcCtx.FileCache.InvalidateUserFileListCache(l.ctx, in.UserId, parentId); err != nil {
				l.Logger.Errorf("RestoreFiles InvalidateUserFileListCache error: %v", err)
			}
		}
	}

	return &pb.RestoreFilesResp{
		RestoredCount: restoredCount,
		UsedSize:      usedSize,
	}, nil
}

// restoreFolder 递归恢复文件夹内容
func (l *RestoreFilesLogic) restoreFolder(folderId uint64, userId int64) (usedSize uint64, restoredCount int64) {
	// 查询文件夹下已删除的文件
	// 这里简化处理，实际应该根据 parent_id 查询
	// 由于软删除时子文件也会被删除，这里不需要递归恢复
	// 子文件会随着父文件夹一起被软删除，但它们的 parent_id 仍然指向父文件夹
	// 所以恢复父文件夹后，子文件需要单独恢复

	// 暂时不递归恢复子文件，用户需要单独恢复
	return 0, 0
}

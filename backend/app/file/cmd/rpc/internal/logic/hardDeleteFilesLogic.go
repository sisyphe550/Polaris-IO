package logic

import (
	"context"
	"errors"

	"polaris-io/backend/app/file/cmd/rpc/internal/svc"
	"polaris-io/backend/app/file/cmd/rpc/pb"

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
			// 减少 MongoDB file_meta 的引用计数
			if file.Hash != "" {
				if err := l.svcCtx.FileMetaModel.DecrRefCount(l.ctx, file.Hash, 1); err != nil {
					l.Logger.Errorf("HardDeleteFiles DecrRefCount error: %v", err)
					// 继续处理，不影响删除
				}

				// TODO: 如果引用计数为 0，删除 S3 文件
				// 这里可以发送一个异步任务到 mqueue 来处理
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
			// 减少引用计数
			if file.Hash != "" {
				if err := l.svcCtx.FileMetaModel.DecrRefCount(l.ctx, file.Hash, 1); err != nil {
					l.Logger.Errorf("hardDeleteFolder DecrRefCount error: %v", err)
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

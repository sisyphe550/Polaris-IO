package logic

import (
	"context"
	"errors"

	"polaris-io/backend/app/file/cmd/rpc/internal/svc"
	"polaris-io/backend/app/file/cmd/rpc/pb"
	"polaris-io/backend/app/file/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type SoftDeleteFilesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSoftDeleteFilesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SoftDeleteFilesLogic {
	return &SoftDeleteFilesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SoftDeleteFiles 软删除文件/文件夹 (移入回收站)
func (l *SoftDeleteFilesLogic) SoftDeleteFiles(in *pb.SoftDeleteFilesReq) (*pb.SoftDeleteFilesResp, error) {
	if in.UserId == 0 || len(in.Identities) == 0 {
		return nil, errors.New("userId and identities are required")
	}

	var deletedCount int64
	var freedSize uint64
	affectedParentIds := make(map[int64]struct{}) // 记录受影响的父目录

	for _, identity := range in.Identities {
		// 查询文件
		file, err := l.svcCtx.UserRepositoryModel.FindOneByIdentity(l.ctx, identity)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				continue // 跳过不存在的文件
			}
			l.Logger.Errorf("SoftDeleteFiles FindOneByIdentity error: %v", err)
			continue
		}

		// 权限验证
		if int64(file.UserId) != in.UserId {
			continue
		}

		// 记录受影响的父目录
		affectedParentIds[int64(file.ParentId)] = struct{}{}

		// 如果是文件夹，需要递归删除所有子文件/文件夹
		if file.Ext == "" && file.Hash == "" {
			subFreedSize, subDeletedCount := l.deleteFolder(file.Id, in.UserId)
			freedSize += subFreedSize
			deletedCount += subDeletedCount
		}

		// 软删除当前文件/文件夹
		if err := l.svcCtx.UserRepositoryModel.DeleteSoft(l.ctx, nil, file); err != nil {
			l.Logger.Errorf("SoftDeleteFiles DeleteSoft error: %v", err)
			continue
		}

		deletedCount++
		freedSize += file.Size

		// 发送 Kafka 删除事件
		if err := l.svcCtx.KafkaProducer.SendFileDeleted(l.ctx, in.UserId, int64(file.Id), identity); err != nil {
			l.Logger.Errorf("SoftDeleteFiles SendFileDeleted error: %v", err)
		}
	}

	// 清除文件列表缓存
	if l.svcCtx.FileCache != nil && deletedCount > 0 {
		for parentId := range affectedParentIds {
			if err := l.svcCtx.FileCache.InvalidateUserFileListCache(l.ctx, in.UserId, parentId); err != nil {
				l.Logger.Errorf("SoftDeleteFiles InvalidateUserFileListCache error: %v", err)
			}
		}
	}

	return &pb.SoftDeleteFilesResp{
		DeletedCount: deletedCount,
		FreedSize:    freedSize,
	}, nil
}

// deleteFolder 递归删除文件夹内容
func (l *SoftDeleteFilesLogic) deleteFolder(folderId uint64, userId int64) (freedSize uint64, deletedCount int64) {
	// 查询文件夹下的所有文件和子文件夹
	builder := l.svcCtx.UserRepositoryModel.SelectBuilder().
		Where("user_id = ?", userId).
		Where("parent_id = ?", folderId)

	children, err := l.svcCtx.UserRepositoryModel.FindAll(l.ctx, builder, "")
	if err != nil {
		l.Logger.Errorf("deleteFolder FindAll error: %v", err)
		return 0, 0
	}

	for _, child := range children {
		// 如果是文件夹，递归删除
		if child.Ext == "" && child.Hash == "" {
			subFreedSize, subDeletedCount := l.deleteFolder(child.Id, userId)
			freedSize += subFreedSize
			deletedCount += subDeletedCount
		}

		// 软删除
		if err := l.svcCtx.UserRepositoryModel.DeleteSoft(l.ctx, nil, child); err != nil {
			l.Logger.Errorf("deleteFolder DeleteSoft error: %v", err)
			continue
		}

		deletedCount++
		freedSize += child.Size
	}

	return freedSize, deletedCount
}

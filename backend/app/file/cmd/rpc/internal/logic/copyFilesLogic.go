package logic

import (
	"context"
	"errors"

	"polaris-io/backend/app/file/cmd/rpc/internal/svc"
	"polaris-io/backend/app/file/cmd/rpc/pb"
	"polaris-io/backend/app/file/model"
	"polaris-io/backend/pkg/xerr"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type CopyFilesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCopyFilesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CopyFilesLogic {
	return &CopyFilesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CopyFiles 复制文件/文件夹
func (l *CopyFilesLogic) CopyFiles(in *pb.CopyFilesReq) (*pb.CopyFilesResp, error) {
	if in.UserId == 0 || len(in.Identities) == 0 {
		return nil, errors.New("userId and identities are required")
	}

	// 检查目标目录是否存在（如果不是根目录）
	if in.TargetId > 0 {
		targetFile, err := l.svcCtx.UserRepositoryModel.FindOne(l.ctx, uint64(in.TargetId))
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil, xerr.NewErrCode(xerr.FILE_PARENT_NOT_EXIST)
			}
			return nil, err
		}
		// 检查目标是文件夹
		if targetFile.Ext != "" || targetFile.Hash != "" {
			return nil, xerr.NewErrCode(xerr.FILE_PARENT_NOT_EXIST)
		}
		// 检查权限
		if int64(targetFile.UserId) != in.UserId {
			return nil, xerr.NewErrCode(xerr.FILE_PARENT_NOT_EXIST)
		}
	}

	var copiedCount int64

	for _, identity := range in.Identities {
		// 查询源文件
		file, err := l.svcCtx.UserRepositoryModel.FindOneByIdentity(l.ctx, identity)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				continue
			}
			l.Logger.Errorf("CopyFiles FindOneByIdentity error: %v", err)
			continue
		}

		// 权限验证
		if int64(file.UserId) != in.UserId {
			continue
		}

		// 如果是文件夹，需要递归复制
		if file.Ext == "" && file.Hash == "" {
			count := l.copyFolder(file, in.TargetId, in.UserId)
			copiedCount += count
		} else {
			// 复制文件
			if err := l.copyFile(file, in.TargetId, in.UserId); err != nil {
				l.Logger.Errorf("CopyFiles copyFile error: %v", err)
				continue
			}
			copiedCount++
		}
	}

	return &pb.CopyFilesResp{
		CopiedCount: copiedCount,
	}, nil
}

// copyFile 复制单个文件
func (l *CopyFilesLogic) copyFile(src *model.UserRepository, targetId int64, userId int64) error {
	// 增加 MongoDB file_meta 引用计数
	if src.Hash != "" {
		if err := l.svcCtx.FileMetaModel.IncrRefCount(l.ctx, src.Hash, 1); err != nil {
			l.Logger.Errorf("copyFile IncrRefCount error: %v", err)
			// 不影响复制，继续
		}
	}

	// 创建新的文件记录
	newFile := &model.UserRepository{
		Identity: uuid.New().String(),
		Hash:     src.Hash,
		UserId:   uint64(userId),
		ParentId: uint64(targetId),
		Name:     src.Name,
		Ext:      src.Ext,
		Size:     src.Size,
		Path:     src.Path,
	}

	_, err := l.svcCtx.UserRepositoryModel.Insert(l.ctx, nil, newFile)
	return err
}

// copyFolder 递归复制文件夹
func (l *CopyFilesLogic) copyFolder(src *model.UserRepository, targetId int64, userId int64) int64 {
	// 创建新文件夹
	newFolder := &model.UserRepository{
		Identity: uuid.New().String(),
		Hash:     "",
		UserId:   uint64(userId),
		ParentId: uint64(targetId),
		Name:     src.Name,
		Ext:      "",
		Size:     0,
		Path:     "",
	}

	result, err := l.svcCtx.UserRepositoryModel.Insert(l.ctx, nil, newFolder)
	if err != nil {
		l.Logger.Errorf("copyFolder Insert error: %v", err)
		return 0
	}

	newFolderId, _ := result.LastInsertId()
	var copiedCount int64 = 1

	// 查询子文件/文件夹
	builder := l.svcCtx.UserRepositoryModel.SelectBuilder().
		Where("user_id = ?", userId).
		Where("parent_id = ?", src.Id)

	children, err := l.svcCtx.UserRepositoryModel.FindAll(l.ctx, builder, "")
	if err != nil {
		l.Logger.Errorf("copyFolder FindAll error: %v", err)
		return copiedCount
	}

	// 递归复制子项
	for _, child := range children {
		if child.Ext == "" && child.Hash == "" {
			copiedCount += l.copyFolder(child, newFolderId, userId)
		} else {
			if err := l.copyFile(child, newFolderId, userId); err == nil {
				copiedCount++
			}
		}
	}

	return copiedCount
}

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

type CreateFolderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateFolderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateFolderLogic {
	return &CreateFolderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateFolder 创建文件夹
func (l *CreateFolderLogic) CreateFolder(in *pb.CreateFolderReq) (*pb.CreateFolderResp, error) {
	// 参数校验
	if in.UserId == 0 || in.Name == "" {
		return nil, errors.New("userId and name are required")
	}

	// 检查父目录是否存在（如果不是根目录）
	if in.ParentId > 0 {
		parentBuilder := l.svcCtx.UserRepositoryModel.SelectBuilder().
			Where("id = ?", in.ParentId).
			Where("user_id = ?", in.UserId).
			Where("ext = ?", "") // 文件夹的 ext 为空
		parents, err := l.svcCtx.UserRepositoryModel.FindAll(l.ctx, parentBuilder, "")
		if err != nil || len(parents) == 0 {
			return nil, xerr.NewErrCode(xerr.FILE_PARENT_NOT_EXIST)
		}
	}

	// 检查同目录下是否有同名文件夹
	existBuilder := l.svcCtx.UserRepositoryModel.SelectBuilder().
		Where("user_id = ?", in.UserId).
		Where("parent_id = ?", in.ParentId).
		Where("name = ?", in.Name).
		Where("ext = ?", "") // 文件夹
	existFiles, err := l.svcCtx.UserRepositoryModel.FindAll(l.ctx, existBuilder, "")
	if err != nil {
		l.Logger.Errorf("CreateFolder check exist error: %v", err)
		return nil, err
	}
	if len(existFiles) > 0 {
		return nil, xerr.NewErrCode(xerr.FOLDER_ALREADY_EXISTS)
	}

	// 生成文件夹唯一标识
	identity := uuid.New().String()

	// 创建文件夹记录
	folder := &model.UserRepository{
		Identity: identity,
		Hash:     "",           // 文件夹没有 hash
		UserId:   uint64(in.UserId),
		ParentId: uint64(in.ParentId),
		Name:     in.Name,
		Ext:      "",           // 文件夹的 ext 为空
		Size:     0,            // 文件夹大小为 0
		Path:     "",           // 文件夹没有 S3 路径
	}

	result, err := l.svcCtx.UserRepositoryModel.Insert(l.ctx, nil, folder)
	if err != nil {
		l.Logger.Errorf("CreateFolder Insert error: %v", err)
		return nil, xerr.NewErrCode(xerr.FOLDER_CREATE_FAILED)
	}

	folderId, _ := result.LastInsertId()

	// 清除文件列表缓存
	if l.svcCtx.FileCache != nil {
		if err := l.svcCtx.FileCache.InvalidateUserFileListCache(l.ctx, in.UserId, in.ParentId); err != nil {
			l.Logger.Errorf("CreateFolder InvalidateUserFileListCache error: %v", err)
		}
	}

	return &pb.CreateFolderResp{
		Identity: identity,
		Id:       folderId,
	}, nil
}

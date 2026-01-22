package logic

import (
	"context"
	"errors"

	"polaris-io/backend/app/file/cmd/rpc/internal/svc"
	"polaris-io/backend/app/file/cmd/rpc/pb"
	"polaris-io/backend/app/file/model"
	"polaris-io/backend/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type RenameFileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRenameFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RenameFileLogic {
	return &RenameFileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RenameFile 重命名文件/文件夹
func (l *RenameFileLogic) RenameFile(in *pb.RenameFileReq) (*pb.RenameFileResp, error) {
	if in.UserId == 0 || in.Identity == "" || in.Name == "" {
		return nil, errors.New("userId, identity and name are required")
	}

	// 查询文件
	file, err := l.svcCtx.UserRepositoryModel.FindOneByIdentity(l.ctx, in.Identity)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, xerr.NewErrCode(xerr.FILE_NOT_EXIST)
		}
		l.Logger.Errorf("RenameFile FindOneByIdentity error: %v", err)
		return nil, err
	}

	// 权限验证
	if int64(file.UserId) != in.UserId {
		return nil, xerr.NewErrCode(xerr.FILE_NOT_EXIST)
	}

	// 检查同目录下是否有同名文件
	isDir := file.Ext == "" && file.Hash == ""
	existBuilder := l.svcCtx.UserRepositoryModel.SelectBuilder().
		Where("user_id = ?", in.UserId).
		Where("parent_id = ?", file.ParentId).
		Where("name = ?", in.Name).
		Where("id != ?", file.Id)
	if isDir {
		existBuilder = existBuilder.Where("ext = ?", "")
	}
	existFiles, err := l.svcCtx.UserRepositoryModel.FindAll(l.ctx, existBuilder, "")
	if err != nil {
		l.Logger.Errorf("RenameFile check exist error: %v", err)
		return nil, err
	}
	if len(existFiles) > 0 {
		return nil, xerr.NewErrCode(xerr.FILE_NAME_DUPLICATE)
	}

	// 更新名称
	file.Name = in.Name
	_, err = l.svcCtx.UserRepositoryModel.Update(l.ctx, nil, file)
	if err != nil {
		l.Logger.Errorf("RenameFile Update error: %v", err)
		return nil, xerr.NewErrCode(xerr.FILE_RENAME_FAILED)
	}

	// 清除文件列表缓存
	if l.svcCtx.FileCache != nil {
		if err := l.svcCtx.FileCache.InvalidateUserFileListCache(l.ctx, in.UserId, int64(file.ParentId)); err != nil {
			l.Logger.Errorf("RenameFile InvalidateUserFileListCache error: %v", err)
		}
	}

	// 发送 Kafka 更新事件
	if err := l.svcCtx.KafkaProducer.SendFileUpdated(
		l.ctx, in.UserId, int64(file.Id), in.Identity, in.Name, int64(file.ParentId)); err != nil {
		l.Logger.Errorf("RenameFile SendFileUpdated error: %v", err)
	}

	return &pb.RenameFileResp{}, nil
}

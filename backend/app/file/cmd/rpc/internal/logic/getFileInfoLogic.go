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

type GetFileInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFileInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFileInfoLogic {
	return &GetFileInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetFileInfo 获取文件信息
func (l *GetFileInfoLogic) GetFileInfo(in *pb.GetFileInfoReq) (*pb.GetFileInfoResp, error) {
	if in.Identity == "" {
		return nil, errors.New("identity is required")
	}

	// 根据 identity 查询
	file, err := l.svcCtx.UserRepositoryModel.FindOneByIdentity(l.ctx, in.Identity)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, xerr.NewErrCode(xerr.FILE_NOT_EXIST)
		}
		l.Logger.Errorf("GetFileInfo FindOneByIdentity error: %v", err)
		return nil, err
	}

	// 权限验证（如果提供了 userId）
	if in.UserId > 0 && int64(file.UserId) != in.UserId {
		return nil, xerr.NewErrCode(xerr.FILE_NOT_EXIST) // 不暴露文件存在信息
	}

	isDir := file.Ext == "" && file.Hash == ""

	return &pb.GetFileInfoResp{
		File: &pb.FileInfo{
			Id:         int64(file.Id),
			Identity:   file.Identity,
			Hash:       file.Hash,
			UserId:     int64(file.UserId),
			ParentId:   int64(file.ParentId),
			Name:       file.Name,
			Ext:        file.Ext,
			Size:       file.Size,
			Path:       file.Path,
			IsDir:      isDir,
			CreateTime: file.CreateTime.Unix(),
			UpdateTime: file.UpdateTime.Unix(),
		},
	}, nil
}

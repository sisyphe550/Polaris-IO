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

type GetFilePathLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFilePathLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFilePathLogic {
	return &GetFilePathLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetFilePath 获取文件路径 (面包屑导航)
func (l *GetFilePathLogic) GetFilePath(in *pb.GetFilePathReq) (*pb.GetFilePathResp, error) {
	if in.Identity == "" {
		return nil, errors.New("identity is required")
	}

	// 查询当前文件
	file, err := l.svcCtx.UserRepositoryModel.FindOneByIdentity(l.ctx, in.Identity)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, xerr.NewErrCode(xerr.FILE_NOT_EXIST)
		}
		l.Logger.Errorf("GetFilePath FindOneByIdentity error: %v", err)
		return nil, err
	}

	// 权限验证
	if in.UserId > 0 && int64(file.UserId) != in.UserId {
		return nil, xerr.NewErrCode(xerr.FILE_NOT_EXIST)
	}

	// 构建路径（从当前文件向上遍历到根目录）
	path := make([]*pb.FileInfo, 0)

	// 添加当前文件
	isDir := file.Ext == "" && file.Hash == ""
	path = append(path, &pb.FileInfo{
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
	})

	// 向上遍历父目录
	parentId := file.ParentId
	for parentId > 0 {
		parent, err := l.svcCtx.UserRepositoryModel.FindOne(l.ctx, parentId)
		if err != nil {
			break
		}

		isParentDir := parent.Ext == "" && parent.Hash == ""
		path = append([]*pb.FileInfo{{
			Id:         int64(parent.Id),
			Identity:   parent.Identity,
			Hash:       parent.Hash,
			UserId:     int64(parent.UserId),
			ParentId:   int64(parent.ParentId),
			Name:       parent.Name,
			Ext:        parent.Ext,
			Size:       parent.Size,
			Path:       parent.Path,
			IsDir:      isParentDir,
			CreateTime: parent.CreateTime.Unix(),
			UpdateTime: parent.UpdateTime.Unix(),
		}}, path...)

		parentId = parent.ParentId
	}

	return &pb.GetFilePathResp{Path: path}, nil
}

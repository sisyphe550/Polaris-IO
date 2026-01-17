package file

import (
	"context"

	"polaris-io/backend/app/file/cmd/api/internal/svc"
	"polaris-io/backend/app/file/cmd/api/internal/types"
	"polaris-io/backend/app/file/cmd/rpc/fileservice"
	"polaris-io/backend/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type FileDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFileDetailLogic 获取文件详情
func NewFileDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileDetailLogic {
	return &FileDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FileDetailLogic) FileDetail(req *types.FileDetailReq) (resp *types.FileDetailResp, err error) {
	userId := ctxdata.GetUidFromCtx(l.ctx)

	// 获取文件信息
	fileResp, err := l.svcCtx.FileRpc.GetFileInfo(l.ctx, &fileservice.GetFileInfoReq{
		Identity: req.Identity,
		UserId:   userId,
	})
	if err != nil {
		return nil, err
	}

	// 获取路径
	pathResp, err := l.svcCtx.FileRpc.GetFilePath(l.ctx, &fileservice.GetFilePathReq{
		Identity: req.Identity,
		UserId:   userId,
	})
	if err != nil {
		return nil, err
	}

	// 转换文件信息
	file := types.FileItem{
		Id:         fileResp.File.Id,
		Identity:   fileResp.File.Identity,
		Name:       fileResp.File.Name,
		Ext:        fileResp.File.Ext,
		Size:       fileResp.File.Size,
		IsDir:      fileResp.File.IsDir,
		ParentId:   fileResp.File.ParentId,
		CreateTime: fileResp.File.CreateTime,
		UpdateTime: fileResp.File.UpdateTime,
	}

	// 转换路径
	path := make([]types.FileItem, 0, len(pathResp.Path))
	for _, p := range pathResp.Path {
		path = append(path, types.FileItem{
			Id:         p.Id,
			Identity:   p.Identity,
			Name:       p.Name,
			Ext:        p.Ext,
			Size:       p.Size,
			IsDir:      p.IsDir,
			ParentId:   p.ParentId,
			CreateTime: p.CreateTime,
			UpdateTime: p.UpdateTime,
		})
	}

	return &types.FileDetailResp{
		File: file,
		Path: path,
	}, nil
}

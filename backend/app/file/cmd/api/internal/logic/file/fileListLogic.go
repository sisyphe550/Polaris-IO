package file

import (
	"context"

	"polaris-io/backend/app/file/cmd/api/internal/svc"
	"polaris-io/backend/app/file/cmd/api/internal/types"
	"polaris-io/backend/app/file/cmd/rpc/fileservice"
	"polaris-io/backend/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type FileListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFileListLogic 列出目录内容
func NewFileListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileListLogic {
	return &FileListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FileListLogic) FileList(req *types.FileListReq) (resp *types.FileListResp, err error) {
	userId := ctxdata.GetUidFromCtx(l.ctx)

	rpcResp, err := l.svcCtx.FileRpc.ListFiles(l.ctx, &fileservice.ListFilesReq{
		UserId:   userId,
		ParentId: req.ParentId,
		Page:     req.Page,
		PageSize: req.PageSize,
		OrderBy:  req.OrderBy,
		Order:    req.Order,
	})
	if err != nil {
		return nil, err
	}

	// 转换响应
	list := make([]types.FileItem, 0, len(rpcResp.List))
	for _, f := range rpcResp.List {
		list = append(list, types.FileItem{
			Id:         f.Id,
			Identity:   f.Identity,
			Name:       f.Name,
			Ext:        f.Ext,
			Size:       f.Size,
			IsDir:      f.IsDir,
			ParentId:   f.ParentId,
			CreateTime: f.CreateTime,
			UpdateTime: f.UpdateTime,
		})
	}

	return &types.FileListResp{
		List:  list,
		Total: rpcResp.Total,
	}, nil
}

package search

import (
	"context"

	"polaris-io/backend/app/search/cmd/api/internal/svc"
	"polaris-io/backend/app/search/cmd/api/internal/types"
	"polaris-io/backend/app/search/cmd/rpc/searchservice"
	"polaris-io/backend/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchFilesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 搜索文件
func NewSearchFilesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchFilesLogic {
	return &SearchFilesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchFilesLogic) SearchFiles(req *types.SearchFilesReq) (resp *types.SearchFilesResp, err error) {
	userId := ctxdata.GetUidFromCtx(l.ctx)

	// 构建 RPC 请求
	rpcReq := &searchservice.SearchFilesReq{
		UserId:   userId,
		Keyword:  req.Keyword,
		Page:     int64(req.Page),
		PageSize: int64(req.PageSize),
		SortBy:   req.SortBy,
		SortDesc: req.SortDesc,
		IsDir:    -1, // 默认搜索全部
	}

	// 扩展名过滤
	if len(req.Ext) > 0 {
		rpcReq.Ext = req.Ext
	}

	// 是否文件夹
	if req.IsDir != nil {
		if *req.IsDir {
			rpcReq.IsDir = 1
		} else {
			rpcReq.IsDir = 0
		}
	}

	// 文件大小过滤
	if req.MinSize != nil {
		rpcReq.MinSize = *req.MinSize
	}
	if req.MaxSize != nil {
		rpcReq.MaxSize = *req.MaxSize
	}

	// 调用 RPC
	rpcResp, err := l.svcCtx.SearchRpc.SearchFiles(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}

	// 转换响应
	list := make([]types.FileItem, 0, len(rpcResp.List))
	for _, item := range rpcResp.List {
		list = append(list, types.FileItem{
			Id:         item.Id,
			FileId:     item.FileId,
			Name:       item.Name,
			Ext:        item.Ext,
			Size:       item.Size,
			IsDir:      item.IsDir,
			ParentId:   item.ParentId,
			CreateTime: item.CreateTime,
			UpdateTime: item.UpdateTime,
		})
	}

	return &types.SearchFilesResp{
		Total: rpcResp.Total,
		List:  list,
	}, nil
}

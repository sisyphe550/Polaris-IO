package logic

import (
	"context"

	"polaris-io/backend/app/search/cmd/rpc/internal/svc"
	"polaris-io/backend/app/search/cmd/rpc/pb"
	"polaris-io/backend/app/search/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchFilesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchFilesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchFilesLogic {
	return &SearchFilesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SearchFiles 搜索文件
func (l *SearchFilesLogic) SearchFiles(in *pb.SearchFilesReq) (*pb.SearchFilesResp, error) {
	// 构建搜索选项
	opts := &types.SearchOptions{
		UserID:   in.UserId,
		Keyword:  in.Keyword,
		Page:     int(in.Page),
		PageSize: int(in.PageSize),
		SortBy:   in.SortBy,
		SortDesc: in.SortDesc,
	}

	// 扩展名过滤
	if len(in.Ext) > 0 {
		opts.Ext = in.Ext
	}

	// 是否文件夹
	if in.IsDir >= 0 {
		isDir := in.IsDir == 1
		opts.IsDir = &isDir
	}

	// 文件大小过滤
	if in.MinSize > 0 {
		opts.MinSize = &in.MinSize
	}
	if in.MaxSize > 0 {
		opts.MaxSize = &in.MaxSize
	}

	// 执行搜索
	result, err := l.svcCtx.ESClient.Search(l.ctx, opts)
	if err != nil {
		l.Logger.Errorf("SearchFiles error: %v", err)
		return nil, err
	}

	// 转换结果
	list := make([]*pb.FileItem, 0, len(result.List))
	for _, doc := range result.List {
		list = append(list, &pb.FileItem{
			Id:         doc.ID,
			FileId:     doc.FileID,
			UserId:     doc.UserID,
			Name:       doc.Name,
			Ext:        doc.Ext,
			Size:       doc.Size,
			IsDir:      doc.IsDir,
			ParentId:   doc.ParentID,
			Hash:       doc.Hash,
			CreateTime: doc.CreateTime.Unix(),
			UpdateTime: doc.UpdateTime.Unix(),
		})
	}

	return &pb.SearchFilesResp{
		Total: result.Total,
		List:  list,
	}, nil
}

package logic

import (
	"context"

	"polaris-io/backend/app/file/cmd/rpc/internal/svc"
	"polaris-io/backend/app/file/cmd/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTrashLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTrashLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTrashLogic {
	return &ListTrashLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListTrash 回收站列表
func (l *ListTrashLogic) ListTrash(in *pb.ListTrashReq) (*pb.ListTrashResp, error) {
	// 分页参数
	page := in.Page
	if page < 1 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 查询已删除的文件
	files, total, err := l.svcCtx.UserRepositoryModel.FindTrashList(l.ctx, uint64(in.UserId), page, pageSize)
	if err != nil {
		l.Logger.Errorf("ListTrash FindTrashList error: %v", err)
		return nil, err
	}

	// 转换为响应格式
	list := make([]*pb.FileInfo, 0, len(files))
	for _, f := range files {
		isDir := f.Ext == "" && f.Hash == ""
		list = append(list, &pb.FileInfo{
			Id:         int64(f.Id),
			Identity:   f.Identity,
			Hash:       f.Hash,
			UserId:     int64(f.UserId),
			ParentId:   int64(f.ParentId),
			Name:       f.Name,
			Ext:        f.Ext,
			Size:       f.Size,
			Path:       f.Path,
			IsDir:      isDir,
			CreateTime: f.CreateTime.Unix(),
			UpdateTime: int64(f.DeleteTime), // 使用 DeleteTime 作为删除时间返回
		})
	}

	return &pb.ListTrashResp{
		List:  list,
		Total: total,
	}, nil
}

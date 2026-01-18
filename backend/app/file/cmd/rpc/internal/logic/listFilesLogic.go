package logic

import (
	"context"

	"polaris-io/backend/app/file/cmd/rpc/internal/svc"
	"polaris-io/backend/app/file/cmd/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFilesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListFilesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFilesLogic {
	return &ListFilesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListFiles 列出目录内容
func (l *ListFilesLogic) ListFiles(in *pb.ListFilesReq) (*pb.ListFilesResp, error) {
	// 构建查询条件
	builder := l.svcCtx.UserRepositoryModel.SelectBuilder().
		Where("user_id = ?", in.UserId).
		Where("parent_id = ?", in.ParentId)

	// 排序
	orderBy := "id DESC" // 默认按 ID 倒序
	if in.OrderBy != "" {
		direction := "ASC"
		if in.Order == "desc" {
			direction = "DESC"
		}
		switch in.OrderBy {
		case "name":
			orderBy = "name " + direction
		case "size":
			orderBy = "size " + direction
		case "create_time":
			orderBy = "create_time " + direction
		case "update_time":
			orderBy = "update_time " + direction
		}
	}

	// 分页参数
	page := in.Page
	if page < 1 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 查询带分页和总数
	files, total, err := l.svcCtx.UserRepositoryModel.FindPageListByPageWithTotal(
		l.ctx, builder, page, pageSize, orderBy)
	if err != nil {
		l.Logger.Errorf("ListFiles FindPageListByPageWithTotal error: %v", err)
		return nil, err
	}

	// 转换为响应格式
	list := make([]*pb.FileInfo, 0, len(files))
	for _, f := range files {
		isDir := f.Ext == "" && f.Hash == "" // 文件夹的 ext 和 hash 都为空
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
			UpdateTime: f.UpdateTime.Unix(),
		})
	}

	return &pb.ListFilesResp{
		List:  list,
		Total: total,
	}, nil
}

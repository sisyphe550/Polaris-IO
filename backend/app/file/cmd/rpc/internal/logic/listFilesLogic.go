package logic

import (
	"context"

	"polaris-io/backend/app/file/cmd/rpc/internal/svc"
	"polaris-io/backend/app/file/cmd/rpc/pb"
	"polaris-io/backend/pkg/filecache"

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
// 使用 Redis 缓存加速查询（5分钟过期）
func (l *ListFilesLogic) ListFiles(in *pb.ListFilesReq) (*pb.ListFilesResp, error) {
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

	// 1. 先查 Redis 缓存
	if l.svcCtx.FileCache != nil {
		cached, exists, err := l.svcCtx.FileCache.GetFileList(l.ctx, in.UserId, in.ParentId, page, pageSize, orderBy)
		if err != nil {
			l.Logger.Errorf("ListFiles cache get error: %v", err)
			// 缓存出错，降级查数据库
		} else if exists {
			// 缓存命中
			l.Logger.Infof("ListFiles cache hit: userId=%d, parentId=%d", in.UserId, in.ParentId)
			list := make([]*pb.FileInfo, 0, len(cached.List))
			for _, f := range cached.List {
				list = append(list, &pb.FileInfo{
					Id:         f.Id,
					Identity:   f.Identity,
					Hash:       f.Hash,
					UserId:     f.UserId,
					ParentId:   f.ParentId,
					Name:       f.Name,
					Ext:        f.Ext,
					Size:       f.Size,
					Path:       f.Path,
					IsDir:      f.IsDir,
					CreateTime: f.CreateTime,
					UpdateTime: f.UpdateTime,
				})
			}
			return &pb.ListFilesResp{
				List:  list,
				Total: cached.Total,
			}, nil
		}
	}

	// 2. 缓存未命中，查数据库
	l.Logger.Infof("ListFiles cache miss, querying database: userId=%d, parentId=%d", in.UserId, in.ParentId)

	// 构建查询条件
	builder := l.svcCtx.UserRepositoryModel.SelectBuilder().
		Where("user_id = ?", in.UserId).
		Where("parent_id = ?", in.ParentId)

	// 查询带分页和总数
	files, total, err := l.svcCtx.UserRepositoryModel.FindPageListByPageWithTotal(
		l.ctx, builder, page, pageSize, orderBy)
	if err != nil {
		l.Logger.Errorf("ListFiles FindPageListByPageWithTotal error: %v", err)
		return nil, err
	}

	// 转换为响应格式
	list := make([]*pb.FileInfo, 0, len(files))
	cacheList := make([]filecache.FileItemCache, 0, len(files))
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
		cacheList = append(cacheList, filecache.FileItemCache{
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

	// 3. 写入缓存（5分钟过期）
	if l.svcCtx.FileCache != nil {
		cacheData := &filecache.FileListCache{
			List:  cacheList,
			Total: total,
		}
		if err := l.svcCtx.FileCache.SetFileList(l.ctx, in.UserId, in.ParentId, page, pageSize, orderBy, cacheData); err != nil {
			l.Logger.Errorf("ListFiles cache set error: %v", err)
			// 缓存写入失败不影响主流程
		}
	}

	return &pb.ListFilesResp{
		List:  list,
		Total: total,
	}, nil
}

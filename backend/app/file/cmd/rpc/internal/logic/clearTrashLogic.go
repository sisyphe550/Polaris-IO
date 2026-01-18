package logic

import (
	"context"
	"errors"

	"polaris-io/backend/app/file/cmd/rpc/internal/svc"
	"polaris-io/backend/app/file/cmd/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClearTrashLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewClearTrashLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearTrashLogic {
	return &ClearTrashLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ClearTrash 清空回收站
func (l *ClearTrashLogic) ClearTrash(in *pb.ClearTrashReq) (*pb.ClearTrashResp, error) {
	if in.UserId == 0 {
		return nil, errors.New("userId is required")
	}

	var deletedCount int64

	// 分批查询并删除，避免一次查询过多数据
	for {
		// 每次查询 100 条
		files, _, err := l.svcCtx.UserRepositoryModel.FindTrashList(l.ctx, uint64(in.UserId), 1, 100)
		if err != nil {
			l.Logger.Errorf("ClearTrash FindTrashList error: %v", err)
			return nil, err
		}

		if len(files) == 0 {
			break
		}

		for _, file := range files {
			// 减少 MongoDB file_meta 的引用计数（只对文件，不对文件夹）
			if file.Hash != "" {
				if err := l.svcCtx.FileMetaModel.DecrRefCount(l.ctx, file.Hash, 1); err != nil {
					l.Logger.Errorf("ClearTrash DecrRefCount error: %v", err)
					// 继续处理
				}
			}

			// 彻底删除 MySQL 记录
			if err := l.svcCtx.UserRepositoryModel.HardDelete(l.ctx, nil, file.Id); err != nil {
				l.Logger.Errorf("ClearTrash HardDelete error: %v", err)
				continue
			}

			deletedCount++
		}
	}

	return &pb.ClearTrashResp{
		DeletedCount: deletedCount,
	}, nil
}

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

	// TODO: 实现清空回收站逻辑
	// 1. 查询所有已删除的文件
	// 2. 批量执行彻底删除

	// 简化实现
	return &pb.ClearTrashResp{
		DeletedCount: 0,
	}, nil
}

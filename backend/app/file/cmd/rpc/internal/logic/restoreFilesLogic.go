package logic

import (
	"context"
	"errors"

	"polaris-io/backend/app/file/cmd/rpc/internal/svc"
	"polaris-io/backend/app/file/cmd/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type RestoreFilesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRestoreFilesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RestoreFilesLogic {
	return &RestoreFilesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RestoreFiles 恢复文件 (从回收站恢复)
func (l *RestoreFilesLogic) RestoreFiles(in *pb.RestoreFilesReq) (*pb.RestoreFilesResp, error) {
	if in.UserId == 0 || len(in.Identities) == 0 {
		return nil, errors.New("userId and identities are required")
	}

	// TODO: 实现恢复逻辑
	// 1. 查询已删除的文件 (del_state = 1)
	// 2. 检查原目录是否存在，不存在则恢复到根目录
	// 3. 更新 del_state = 0, delete_time = 0
	// 4. 返回恢复后占用的空间大小 (用于扣减配额)

	// 简化实现，后续扩展 Model 层
	return &pb.RestoreFilesResp{
		RestoredCount: 0,
		UsedSize:      0,
	}, nil
}

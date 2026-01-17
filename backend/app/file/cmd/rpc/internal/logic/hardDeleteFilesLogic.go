package logic

import (
	"context"
	"errors"

	"polaris-io/backend/app/file/cmd/rpc/internal/svc"
	"polaris-io/backend/app/file/cmd/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type HardDeleteFilesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHardDeleteFilesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HardDeleteFilesLogic {
	return &HardDeleteFilesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// HardDeleteFiles 彻底删除文件
func (l *HardDeleteFilesLogic) HardDeleteFiles(in *pb.HardDeleteFilesReq) (*pb.HardDeleteFilesResp, error) {
	if in.UserId == 0 || len(in.Identities) == 0 {
		return nil, errors.New("userId and identities are required")
	}

	// TODO: 实现彻底删除逻辑
	// 1. 查询已删除的文件 (del_state = 1)
	// 2. 减少 MongoDB file_meta 的引用计数
	// 3. 如果引用计数为 0，删除 S3 文件和 MongoDB 记录
	// 4. 删除 MySQL user_repository 记录

	// 简化实现
	return &pb.HardDeleteFilesResp{
		DeletedCount: 0,
	}, nil
}

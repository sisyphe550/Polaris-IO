package searchservicelogic

import (
	"context"

	"polaris-io/backend/app/search/cmd/rpc/internal/svc"
	"polaris-io/backend/app/search/cmd/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserStatsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserStatsLogic {
	return &GetUserStatsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取用户文件统计
func (l *GetUserStatsLogic) GetUserStats(in *pb.GetUserStatsReq) (*pb.GetUserStatsResp, error) {
	// todo: add your logic here and delete this line

	return &pb.GetUserStatsResp{}, nil
}

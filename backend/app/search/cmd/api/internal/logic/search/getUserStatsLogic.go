package search

import (
	"context"

	"polaris-io/backend/app/search/cmd/api/internal/svc"
	"polaris-io/backend/app/search/cmd/api/internal/types"
	"polaris-io/backend/app/search/cmd/rpc/searchservice"
	"polaris-io/backend/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取用户文件统计
func NewGetUserStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserStatsLogic {
	return &GetUserStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserStatsLogic) GetUserStats(req *types.GetUserStatsReq) (resp *types.GetUserStatsResp, err error) {
	userId := ctxdata.GetUidFromCtx(l.ctx)

	rpcResp, err := l.svcCtx.SearchRpc.GetUserStats(l.ctx, &searchservice.GetUserStatsReq{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}

	return &types.GetUserStatsResp{
		TotalFiles:   rpcResp.TotalFiles,
		TotalFolders: rpcResp.TotalFolders,
	}, nil
}

package logic

import (
	"context"

	"polaris-io/backend/app/search/cmd/rpc/internal/svc"
	"polaris-io/backend/app/search/cmd/rpc/pb"
	"polaris-io/backend/app/search/types"

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

// GetUserStats 获取用户文件统计
func (l *GetUserStatsLogic) GetUserStats(in *pb.GetUserStatsReq) (*pb.GetUserStatsResp, error) {
	// 统计文件数量
	isFile := false
	fileResult, err := l.svcCtx.ESClient.Search(l.ctx, &types.SearchOptions{
		UserID:   in.UserId,
		IsDir:    &isFile,
		Page:     1,
		PageSize: 1,
	})
	if err != nil {
		l.Logger.Errorf("GetUserStats count files error: %v", err)
		return nil, err
	}

	// 统计文件夹数量
	isDir := true
	folderResult, err := l.svcCtx.ESClient.Search(l.ctx, &types.SearchOptions{
		UserID:   in.UserId,
		IsDir:    &isDir,
		Page:     1,
		PageSize: 1,
	})
	if err != nil {
		l.Logger.Errorf("GetUserStats count folders error: %v", err)
		return nil, err
	}

	return &pb.GetUserStatsResp{
		TotalFiles:   fileResult.Total,
		TotalFolders: folderResult.Total,
	}, nil
}

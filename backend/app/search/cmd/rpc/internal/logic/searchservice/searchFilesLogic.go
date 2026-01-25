package searchservicelogic

import (
	"context"

	"polaris-io/backend/app/search/cmd/rpc/internal/svc"
	"polaris-io/backend/app/search/cmd/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchFilesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchFilesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchFilesLogic {
	return &SearchFilesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// -------- 搜索接口 --------
func (l *SearchFilesLogic) SearchFiles(in *pb.SearchFilesReq) (*pb.SearchFilesResp, error) {
	// todo: add your logic here and delete this line

	return &pb.SearchFilesResp{}, nil
}

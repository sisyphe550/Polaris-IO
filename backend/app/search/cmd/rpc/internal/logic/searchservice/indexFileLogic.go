package searchservicelogic

import (
	"context"

	"polaris-io/backend/app/search/cmd/rpc/internal/svc"
	"polaris-io/backend/app/search/cmd/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type IndexFileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIndexFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IndexFileLogic {
	return &IndexFileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// -------- 索引管理（内部调用） --------
func (l *IndexFileLogic) IndexFile(in *pb.IndexFileReq) (*pb.IndexFileResp, error) {
	// todo: add your logic here and delete this line

	return &pb.IndexFileResp{}, nil
}

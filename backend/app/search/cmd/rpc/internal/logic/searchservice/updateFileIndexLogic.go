package searchservicelogic

import (
	"context"

	"polaris-io/backend/app/search/cmd/rpc/internal/svc"
	"polaris-io/backend/app/search/cmd/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateFileIndexLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateFileIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFileIndexLogic {
	return &UpdateFileIndexLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新文件索引
func (l *UpdateFileIndexLogic) UpdateFileIndex(in *pb.UpdateFileIndexReq) (*pb.UpdateFileIndexResp, error) {
	// todo: add your logic here and delete this line

	return &pb.UpdateFileIndexResp{}, nil
}

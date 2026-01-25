package searchservicelogic

import (
	"context"

	"polaris-io/backend/app/search/cmd/rpc/internal/svc"
	"polaris-io/backend/app/search/cmd/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteFileIndexLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteFileIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFileIndexLogic {
	return &DeleteFileIndexLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除文件索引
func (l *DeleteFileIndexLogic) DeleteFileIndex(in *pb.DeleteFileIndexReq) (*pb.DeleteFileIndexResp, error) {
	// todo: add your logic here and delete this line

	return &pb.DeleteFileIndexResp{}, nil
}

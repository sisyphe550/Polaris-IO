package logic

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

// DeleteFileIndex 删除文件索引
func (l *DeleteFileIndexLogic) DeleteFileIndex(in *pb.DeleteFileIndexReq) (*pb.DeleteFileIndexResp, error) {
	if err := l.svcCtx.ESClient.DeleteDocument(l.ctx, in.Id); err != nil {
		l.Logger.Errorf("DeleteFileIndex error: %v", err)
		return nil, err
	}

	l.Logger.Infof("Deleted file index: id=%s", in.Id)
	return &pb.DeleteFileIndexResp{}, nil
}

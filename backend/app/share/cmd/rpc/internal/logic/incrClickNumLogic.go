package logic

import (
	"context"

	"polaris-io/backend/app/share/cmd/rpc/internal/svc"
	"polaris-io/backend/app/share/cmd/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type IncrClickNumLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIncrClickNumLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IncrClickNumLogic {
	return &IncrClickNumLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// IncrClickNum 增加点击次数
func (l *IncrClickNumLogic) IncrClickNum(in *pb.IncrClickNumReq) (*pb.IncrClickNumResp, error) {
	if err := l.svcCtx.ShareModel.IncrClickNum(l.ctx, in.Identity); err != nil {
		l.Logger.Errorf("IncrClickNum error: %v", err)
		return nil, err
	}
	return &pb.IncrClickNumResp{}, nil
}

package logic

import (
	"context"
	"errors"

	"polaris-io/backend/app/share/cmd/rpc/internal/svc"
	"polaris-io/backend/app/share/cmd/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelShareLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelShareLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelShareLogic {
	return &CancelShareLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CancelShare 取消分享
func (l *CancelShareLogic) CancelShare(in *pb.CancelShareReq) (*pb.CancelShareResp, error) {
	if in.UserId == 0 || len(in.Identities) == 0 {
		return nil, errors.New("userId and identities are required")
	}

	// 批量软删除
	cancelledCount, err := l.svcCtx.ShareModel.BatchDeleteSoft(l.ctx, nil, uint64(in.UserId), in.Identities)
	if err != nil {
		l.Logger.Errorf("CancelShare BatchDeleteSoft error: %v", err)
		return nil, err
	}

	return &pb.CancelShareResp{
		CancelledCount: cancelledCount,
	}, nil
}

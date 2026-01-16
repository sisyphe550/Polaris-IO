package logic

import (
	"context"

	"polaris-io/backend/app/user/cmd/rpc/internal/svc"
	"polaris-io/backend/app/user/cmd/rpc/pb"
	"polaris-io/backend/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type RefundQuotaLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefundQuotaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefundQuotaLogic {
	return &RefundQuotaLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RefundQuota 退还配额 (删除文件时调用)
func (l *RefundQuotaLogic) RefundQuota(in *pb.RefundQuotaReq) (*pb.RefundQuotaResp, error) {
	// 退还配额
	err := l.svcCtx.UserQuotaModel.RefundQuota(l.ctx, uint64(in.UserId), in.Size)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "RefundQuota err:%v, userId:%d, size:%d", err, in.UserId, in.Size)
	}

	return &pb.RefundQuotaResp{
		Success: true,
	}, nil
}

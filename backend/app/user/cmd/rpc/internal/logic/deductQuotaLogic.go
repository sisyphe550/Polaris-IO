package logic

import (
	"context"

	"polaris-io/backend/app/user/cmd/rpc/internal/svc"
	"polaris-io/backend/app/user/cmd/rpc/pb"
	"polaris-io/backend/app/user/model"
	"polaris-io/backend/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeductQuotaLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeductQuotaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeductQuotaLogic {
	return &DeductQuotaLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeductQuota 扣减配额 (上传文件前调用，带并发保护)
func (l *DeductQuotaLogic) DeductQuota(in *pb.DeductQuotaReq) (*pb.DeductQuotaResp, error) {
	// 使用原子操作扣减配额，防止并发超额
	err := l.svcCtx.UserQuotaModel.DeductQuota(l.ctx, uint64(in.UserId), in.Size)
	if err != nil {
		if err == model.ErrQuotaExceeded {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.USER_QUOTA_EXCEEDED), "userId:%d size:%d quota exceeded", in.UserId, in.Size)
		}
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "DeductQuota err:%v, userId:%d, size:%d", err, in.UserId, in.Size)
	}

	return &pb.DeductQuotaResp{
		Success: true,
	}, nil
}

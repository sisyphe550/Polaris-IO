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

type GetUserQuotaLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserQuotaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserQuotaLogic {
	return &GetUserQuotaLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetUserQuota 获取用户配额
func (l *GetUserQuotaLogic) GetUserQuota(in *pb.GetUserQuotaReq) (*pb.GetUserQuotaResp, error) {
	// 根据用户 ID 查询配额
	quota, err := l.svcCtx.UserQuotaModel.FindOneByUserId(l.ctx, uint64(in.UserId))
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.USER_QUOTA_NOT_EXIST), "userId:%d quota not found", in.UserId)
		}
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "FindOneByUserId err:%v, userId:%d", err, in.UserId)
	}

	return &pb.GetUserQuotaResp{
		Quota: &pb.UserQuota{
			UserId:    int64(quota.UserId),
			TotalSize: quota.TotalSize,
			UsedSize:  quota.UsedSize,
		},
	}, nil
}

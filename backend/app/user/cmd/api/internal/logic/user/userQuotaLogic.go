package user

import (
	"context"

	"polaris-io/backend/app/user/cmd/api/internal/svc"
	"polaris-io/backend/app/user/cmd/api/internal/types"
	"polaris-io/backend/app/user/cmd/rpc/usercenter"
	"polaris-io/backend/pkg/ctxdata"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type UserQuotaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUserQuotaLogic 获取当前用户配额
func NewUserQuotaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserQuotaLogic {
	return &UserQuotaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserQuotaLogic) UserQuota(req *types.UserQuotaReq) (resp *types.UserQuotaResp, err error) {
	// 从 JWT 中获取用户 ID
	userId := ctxdata.GetUidFromCtx(l.ctx)

	// 调用 RPC 获取用户配额
	quotaResp, err := l.svcCtx.UsercenterRpc.GetUserQuota(l.ctx, &usercenter.GetUserQuotaReq{
		UserId: userId,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "userId: %d", userId)
	}

	return &types.UserQuotaResp{
		Quota: types.UserQuota{
			TotalSize: quotaResp.Quota.TotalSize,
			UsedSize:  quotaResp.Quota.UsedSize,
		},
	}, nil
}

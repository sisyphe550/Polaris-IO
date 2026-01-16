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

type UserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUserInfoLogic 获取当前用户信息
func NewUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserInfoLogic {
	return &UserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserInfoLogic) UserInfo(req *types.UserInfoReq) (resp *types.UserInfoResp, err error) {
	// 从 JWT 中获取用户 ID
	userId := ctxdata.GetUidFromCtx(l.ctx)

	// 调用 RPC 获取用户信息
	userInfoResp, err := l.svcCtx.UsercenterRpc.GetUserInfo(l.ctx, &usercenter.GetUserInfoReq{
		UserId: userId,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "userId: %d", userId)
	}

	return &types.UserInfoResp{
		User: types.User{
			Id:     userInfoResp.User.Id,
			Mobile: userInfoResp.User.Mobile,
			Name:   userInfoResp.User.Name,
			Avatar: userInfoResp.User.Avatar,
			Info:   userInfoResp.User.Info,
		},
	}, nil
}

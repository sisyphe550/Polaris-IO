package user

import (
	"context"

	"polaris-io/backend/app/user/cmd/api/internal/svc"
	"polaris-io/backend/app/user/cmd/api/internal/types"
	"polaris-io/backend/app/user/cmd/rpc/usercenter"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewLoginLogic 用户登录
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	// 调用 RPC 登录服务
	loginResp, err := l.svcCtx.UsercenterRpc.Login(l.ctx, &usercenter.LoginReq{
		Mobile:   req.Mobile,
		Password: req.Password,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "req: %+v", req)
	}

	return &types.LoginResp{
		AccessToken:  loginResp.AccessToken,
		AccessExpire: loginResp.AccessExpire,
		RefreshAfter: loginResp.RefreshAfter,
	}, nil
}

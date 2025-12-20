package user

import (
	"context"

	"shared-board/backend/app/user/cmd/api/internal/svc"
	"shared-board/backend/app/user/cmd/api/internal/types"
	"shared-board/backend/app/user/cmd/rpc/usercenter"

	"github.com/jinzhu/copier"
	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 用户登录
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	// todo: add your logic here and delete this line

	// 1. 调用 RPC 登录接口
	loginResp, err := l.svcCtx.UsercenterRpc.Login(l.ctx, &usercenter.LoginReq{
		AuthType: "mobile",   // 目前仅支持手机号登录
		AuthKey:  req.Mobile, // 手机号作为唯一标识
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	// 2. 组装返回结果
	var res types.LoginResp
	_ = copier.Copy(&res, loginResp)

	return &res, nil
	// return
}

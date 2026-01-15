package logic

import (
	"context"

	"polaris-io/backend/app/user/cmd/rpc/internal/svc"
	"polaris-io/backend/app/user/cmd/rpc/pb"
	"polaris-io/backend/app/user/model"
	"polaris-io/backend/pkg/tool"
	"polaris-io/backend/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Login 用户登录
func (l *LoginLogic) Login(in *pb.LoginReq) (*pb.LoginResp, error) {
	// 1. 根据手机号查询用户
	user, err := l.svcCtx.UserModel.FindOneByMobile(l.ctx, in.Mobile)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.USER_NOT_EXIST), "mobile:%s not found", in.Mobile)
		}
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "FindOneByMobile err:%v, mobile:%s", err, in.Mobile)
	}

	// 2. 校验密码
	if user.Password != tool.Md5ByString(in.Password) {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.USER_PASSWORD_ERROR), "password error, mobile:%s", in.Mobile)
	}

	// 3. 生成 Token
	generateTokenLogic := NewGenerateTokenLogic(l.ctx, l.svcCtx)
	tokenResp, err := generateTokenLogic.GenerateToken(&pb.GenerateTokenReq{
		UserId: int64(user.Id),
	})
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.TOKEN_GENERATE_ERROR), "GenerateToken err:%v, userId:%d", err, user.Id)
	}

	return &pb.LoginResp{
		AccessToken:  tokenResp.AccessToken,
		AccessExpire: tokenResp.AccessExpire,
		RefreshAfter: tokenResp.RefreshAfter,
	}, nil
}

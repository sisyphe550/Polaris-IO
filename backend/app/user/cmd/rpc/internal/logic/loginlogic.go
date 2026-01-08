package logic

import (
	"context"
	"strings"

	"shared-board/backend/app/user/cmd/rpc/internal/svc"
	"shared-board/backend/app/user/cmd/rpc/pb"
	"shared-board/backend/app/user/model"
	"shared-board/backend/pkg/xerr"

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

var ErrUsernamePwdError = xerr.NewErrMsg("账号或密码不正确")
var ErrUserNoExistsError = xerr.NewErrMsg("用户不存在")

// 登录
func (l *LoginLogic) Login(in *pb.LoginReq) (*pb.LoginResp, error) {
	if in == nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.REUQEST_PARAM_ERROR), "empty request")
	}
	authType := strings.TrimSpace(in.AuthType)
	if authType == "" {
		authType = "mobile"
	}
	if authType != "mobile" {
		// 暂不实现微信小程序登录
		return nil, errors.Wrapf(xerr.NewErrCodeMsg(xerr.REUQEST_PARAM_ERROR, "暂不支持该登录类型"), "authType:%s", authType)
	}
	authKey := strings.TrimSpace(in.AuthKey)
	if authKey == "" {
		return nil, errors.Wrapf(xerr.NewErrCodeMsg(xerr.REUQEST_PARAM_ERROR, "authKey 不能为空"), "authKey empty")
	}
	if strings.TrimSpace(in.Password) == "" {
		return nil, errors.Wrapf(xerr.NewErrCodeMsg(xerr.REUQEST_PARAM_ERROR, "password 不能为空"), "password empty")
	}

	passwordHash := md5ByString(in.Password)

	var userId uint64

	// 优先走 DB（如果配置了 DataSource）
	if l.svcCtx.UsersModel != nil {
		u, err := l.svcCtx.UsersModel.FindOneByUsername(l.ctx, authKey)
		if err != nil && err != model.ErrNotFound {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "FindOneByUsername err authKey:%s, err:%v", authKey, err)
		}
		if u == nil {
			return nil, errors.Wrapf(ErrUserNoExistsError, "authKey:%s", authKey)
		}
		if u.Password != passwordHash {
			return nil, errors.Wrap(ErrUsernamePwdError, "password mismatch")
		}
		userId = u.Id
	} else {
		// DB 未配置：用内存仓库
		u, ok := l.svcCtx.MemUsers.FindOneByUsername(authKey)
		if !ok || u == nil {
			return nil, errors.Wrapf(ErrUserNoExistsError, "authKey:%s", authKey)
		}
		if u.Password != passwordHash {
			return nil, errors.Wrap(ErrUsernamePwdError, "password mismatch")
		}
		userId = u.Id
	}

	tokenResp, err := NewGenerateTokenLogic(l.ctx, l.svcCtx).GenerateToken(&pb.GenerateTokenReq{
		UserId: int64(userId),
	})
	if err != nil {
		return nil, err
	}

	return &pb.LoginResp{
		AccessToken:  tokenResp.AccessToken,
		AccessExpire: tokenResp.AccessExpire,
		RefreshAfter: tokenResp.RefreshAfter,
	}, nil
}

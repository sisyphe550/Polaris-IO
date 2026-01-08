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
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

var ErrUserAlreadyRegisterError = xerr.NewErrMsg("用户已注册")

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 注册
func (l *RegisterLogic) Register(in *pb.RegisterReq) (*pb.RegisterResp, error) {
	if in == nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.REUQEST_PARAM_ERROR), "empty request")
	}
	authType := strings.TrimSpace(in.AuthType)
	if authType == "" {
		authType = "mobile"
	}
	if authType != "mobile" {
		// 你暂时不实现微信小程序，这里直接拒绝非 mobile
		return nil, errors.Wrapf(xerr.NewErrCodeMsg(xerr.REUQEST_PARAM_ERROR, "暂不支持该登录类型"), "authType:%s", authType)
	}
	if strings.TrimSpace(in.Mobile) == "" {
		return nil, errors.Wrapf(xerr.NewErrCodeMsg(xerr.REUQEST_PARAM_ERROR, "mobile 不能为空"), "mobile empty")
	}
	if strings.TrimSpace(in.Password) == "" {
		return nil, errors.Wrapf(xerr.NewErrCodeMsg(xerr.REUQEST_PARAM_ERROR, "password 不能为空"), "password empty")
	}

	// users 表只有 username/password/avatar/info
	// 将 “mobile/authKey” 映射到 username（优先使用 authKey）
	username := strings.TrimSpace(in.AuthKey)
	if username == "" {
		username = strings.TrimSpace(in.Mobile)
	}

	passwordHash := md5ByString(in.Password)
	var userId uint64

	// 优先走 DB（如果配置了 DataSource）
	if l.svcCtx.UsersModel != nil {
		u, err := l.svcCtx.UsersModel.FindOneByUsername(l.ctx, username)
		if err != nil && err != model.ErrNotFound {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "FindOneByUsername err username:%s, err:%v", username, err)
		}
		if u != nil {
			return nil, errors.Wrapf(ErrUserAlreadyRegisterError, "username:%s", username)
		}

		if err := l.svcCtx.UsersModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
			user := &model.Users{
				Username: username,
				Password: passwordHash,
				Avatar:   "",
				Info:     "",
				Version:  1,
			}
			insertResult, err := l.svcCtx.UsersModel.Insert(ctx, session, user)
			if err != nil {
				return errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "Insert user err:%v", err)
			}
			lastId, err := insertResult.LastInsertId()
			if err != nil {
				return errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "LastInsertId err:%v", err)
			}
			userId = uint64(lastId)
			return nil
		}); err != nil {
			return nil, err
		}
	} else {
		// DB 未配置：使用内存仓库，方便你继续联调 api
		if _, ok := l.svcCtx.MemUsers.FindOneByUsername(username); ok {
			return nil, errors.Wrapf(ErrUserAlreadyRegisterError, "username:%s", username)
		}
		userId = l.svcCtx.MemUsers.Insert(username, passwordHash)
	}

	tokenResp, err := NewGenerateTokenLogic(l.ctx, l.svcCtx).GenerateToken(&pb.GenerateTokenReq{
		UserId: int64(userId),
	})
	if err != nil {
		return nil, err
	}

	return &pb.RegisterResp{
		AccessToken:  tokenResp.AccessToken,
		AccessExpire: tokenResp.AccessExpire,
		RefreshAfter: tokenResp.RefreshAfter,
	}, nil
}

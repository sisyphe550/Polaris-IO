package logic

import (
	"context"
	"time"

	"shared-board/backend/app/user/cmd/rpc/internal/svc"
	"shared-board/backend/app/user/cmd/rpc/pb"
	"shared-board/backend/pkg/ctxdata"
	"shared-board/backend/pkg/xerr"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGenerateTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateTokenLogic {
	return &GenerateTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 生成 Token (核心鉴权逻辑)
func (l *GenerateTokenLogic) GenerateToken(in *pb.GenerateTokenReq) (*pb.GenerateTokenResp, error) {
	if in == nil || in.UserId <= 0 {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.REUQEST_PARAM_ERROR), "invalid userId")
	}

	secret := l.svcCtx.Config.JwtAuth.AccessSecret
	if secret == "" {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.TOKEN_GENERATE_ERROR), "JwtAuth.AccessSecret is empty")
	}

	now := time.Now().Unix()
	accessExpire := l.svcCtx.Config.JwtAuth.AccessExpire
	if accessExpire <= 0 {
		accessExpire = 31536000 // 1 year, keep consistent with your yaml default
	}

	accessToken, err := l.getJwtToken(secret, now, accessExpire, in.UserId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.TOKEN_GENERATE_ERROR), "getJwtToken err userId:%d , err:%v", in.UserId, err)
	}

	return &pb.GenerateTokenResp{
		AccessToken:  accessToken,
		AccessExpire: now + accessExpire,
		RefreshAfter: now + accessExpire/2,
	}, nil
}

func (l *GenerateTokenLogic) getJwtToken(secretKey string, iat, seconds, userId int64) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	claims[ctxdata.CtxKeyJwtUserId] = userId

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}

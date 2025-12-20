package user

import (
	"context"

	"shared-board/backend/app/user/cmd/api/internal/svc"
	"shared-board/backend/app/user/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WxMiniAuthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 微信小程序授权登录(通常需要Token关联)
func NewWxMiniAuthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WxMiniAuthLogic {
	return &WxMiniAuthLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WxMiniAuthLogic) WxMiniAuth(req *types.WxMiniAuthReq) (resp *types.WxMiniAuthResp, err error) {
	// todo: add your logic here and delete this line

	return
}

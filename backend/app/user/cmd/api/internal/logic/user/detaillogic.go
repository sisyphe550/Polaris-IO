package user

import (
	"context"

	"shared-board/backend/app/user/cmd/api/internal/svc"
	"shared-board/backend/app/user/cmd/api/internal/types"
	"shared-board/backend/app/user/cmd/rpc/usercenter"
	"shared-board/backend/pkg/ctxdata"

	"github.com/jinzhu/copier"
	"github.com/zeromicro/go-zero/core/logx"
)

type DetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取当前用户信息
func NewDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DetailLogic {
	return &DetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DetailLogic) Detail(req *types.UserInfoReq) (resp *types.UserInfoResp, err error) {
	// todo: add your logic here and delete this line

	// 1. 从 Context 中获取当前登录用户的 ID
	// (这是由 JWT 中间件解析后放入 Context 的)
	userId := ctxdata.GetUidFromCtx(l.ctx)

	// 2. 调用 RPC 获取用户信息
	userInfoResp, err := l.svcCtx.UsercenterRpc.GetUserInfo(l.ctx, &usercenter.GetUserInfoReq{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}

	// 3. 组装数据返回
	var res types.UserInfoResp
	if userInfoResp.User != nil {
		_ = copier.Copy(&res.User, userInfoResp.User)
	}

	return &res, nil
	// return
}

package user

import (
	"context"

	"shared-board/backend/app/user/cmd/api/internal/svc"
	"shared-board/backend/app/user/cmd/api/internal/types"
	"shared-board/backend/app/user/cmd/rpc/usercenter"

	"github.com/jinzhu/copier"
	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 用户注册
func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.RegisterResp, err error) {
	// todo: add your logic here and delete this line

	// 1. 调用 RPC 服务的 Register 接口
	// 这里将 API 的请求参数转换为 RPC 需要的请求参数
	registerResp, err := l.svcCtx.UsercenterRpc.Register(l.ctx, &usercenter.RegisterReq{
		Mobile:   req.Mobile,
		Password: req.Password,
		AuthKey:  req.Mobile, // 默认使用手机号作为 AuthKey
		AuthType: "mobile",   // 认证类型标记为手机
	})
	if err != nil {
		return nil, err
	}

	// 2. 将 RPC 的返回结果转换为 API 的响应格式
	// 使用 copier 库可以简化结构体赋值 (go get github.com/jinzhu/copier)
	// 如果不想引入 copier，也可以手动逐个字段赋值
	var res types.RegisterResp
	_ = copier.Copy(&res, registerResp)

	return &res, nil
	// return
}

package svc

import (
	"shared-board/backend/app/user/cmd/api/internal/config"
	"shared-board/backend/app/user/cmd/rpc/usercenter" // 引入 RPC 客户端

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config        config.Config
	UsercenterRpc usercenter.Usercenter // 保留 RPC 客户端
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:        c,
		UsercenterRpc: usercenter.NewUsercenter(zrpc.MustNewClient(c.UsercenterRpcConf)), // 连接 RPC
	}
}

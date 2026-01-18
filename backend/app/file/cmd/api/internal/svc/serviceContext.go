package svc

import (
	"polaris-io/backend/app/file/cmd/api/internal/config"
	"polaris-io/backend/app/file/cmd/rpc/fileservice"
	"polaris-io/backend/app/user/cmd/rpc/usercenter"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	// File RPC Client
	FileRpc fileservice.FileService

	// Usercenter RPC Client (用于配额操作)
	UsercenterRpc usercenter.Usercenter
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,

		// File RPC Client
		FileRpc: fileservice.NewFileService(zrpc.MustNewClient(c.FileRpcConf)),

		// Usercenter RPC Client
		UsercenterRpc: usercenter.NewUsercenter(zrpc.MustNewClient(c.UsercenterRpcConf)),
	}
}

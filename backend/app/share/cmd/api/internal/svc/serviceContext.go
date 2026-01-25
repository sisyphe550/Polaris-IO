package svc

import (
	"polaris-io/backend/app/file/cmd/rpc/fileservice"
	"polaris-io/backend/app/share/cmd/api/internal/config"
	"polaris-io/backend/app/share/cmd/rpc/shareservice"
	"polaris-io/backend/app/user/cmd/rpc/usercenter"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	// Share RPC Client
	ShareRpc shareservice.ShareService

	// File RPC Client（用于获取文件信息）
	FileRpc fileservice.FileService

	// Usercenter RPC Client (用于获取用户信息)
	UsercenterRpc usercenter.Usercenter
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,

		// Share RPC Client
		ShareRpc: shareservice.NewShareService(zrpc.MustNewClient(c.ShareRpcConf)),

		// File RPC Client
		FileRpc: fileservice.NewFileService(zrpc.MustNewClient(c.FileRpcConf)),

		// Usercenter RPC Client
		UsercenterRpc: usercenter.NewUsercenter(zrpc.MustNewClient(c.UsercenterRpcConf)),
	}
}

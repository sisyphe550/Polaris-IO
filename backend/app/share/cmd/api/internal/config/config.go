package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf

	// JWT 配置
	JwtAuth struct {
		AccessSecret string
		AccessExpire int64
	}

	// Share RPC 客户端配置
	ShareRpcConf zrpc.RpcClientConf

	// File RPC 客户端配置（用于获取文件信息）
	FileRpcConf zrpc.RpcClientConf

	// Usercenter RPC 客户端配置（用于获取用户信息）
	UsercenterRpcConf zrpc.RpcClientConf

	// 分享配置
	Share struct {
		BaseUrl string // 分享链接前缀
	}
}

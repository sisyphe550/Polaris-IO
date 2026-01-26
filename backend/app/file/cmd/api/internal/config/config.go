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

	// File RPC 客户端配置
	FileRpcConf zrpc.RpcClientConf

	// Usercenter RPC 客户端配置
	UsercenterRpcConf zrpc.RpcClientConf

	// Asynq 配置（异步任务队列）
	Asynq struct {
		Addr     string
		Password string
	}
}

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	JwtAuth struct {
		AccessSecret string
		AccessExpire int64
	}
	// DB struct {
	// 	DataSource string
	// }
	// 必须有这一行，用来接收 yaml 中的 Redis 配置
	// CacheRedis cache.CacheConf

	// 确保有 RPC 客户端配置
	UsercenterRpcConf zrpc.RpcClientConf
}

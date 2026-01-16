package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf // 内部已包含 Redis 配置

	// JWT 配置
	JwtAuth struct {
		AccessSecret string
		AccessExpire int64
	}

	// 数据库配置
	DB struct {
		DataSource string
	}

	// 缓存配置 (用于 sqlc 缓存)
	Cache cache.CacheConf
}

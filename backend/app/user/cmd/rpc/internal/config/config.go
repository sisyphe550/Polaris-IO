package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	// JWT 配置
	JwtAuth struct {
		AccessSecret string
		AccessExpire int64
	}

	// 数据库配置
	DB struct {
		DataSource string
	}

	// Redis 配置
	Redis struct {
		Host string
		Type string
		Pass string
	}

	// 缓存配置
	Cache cache.CacheConf
}

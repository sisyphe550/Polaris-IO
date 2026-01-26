package config

import (
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	service.ServiceConf

	// Redis 配置（Asynq 后端）
	Redis struct {
		Addr     string
		Password string
	}

	// S3 配置
	S3 struct {
		Endpoint  string
		Region    string
		Bucket    string
		AccessKey string
		SecretKey string
		UseSSL    bool
	}

	// MongoDB 配置
	MongoDB struct {
		Uri      string
		Database string
	}

	// MySQL 配置（File 数据库: polaris_file）
	DB struct {
		DataSource string
	}

	// Share MySQL 配置（Share 数据库: polaris_share）
	ShareDB struct {
		DataSource string
	}

	// 缓存配置
	Cache cache.CacheConf

	// Usercenter RPC 客户端配置
	UsercenterRpcConf zrpc.RpcClientConf
}

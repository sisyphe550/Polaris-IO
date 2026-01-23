package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	// 数据库配置
	DB struct {
		DataSource string
	}

	// 缓存配置
	Cache cache.CacheConf

	// MongoDB 配置
	MongoDB struct {
		Uri      string
		Database string
	}

	// S3 (Garage) 配置
	S3 struct {
		Endpoint  string
		Region    string
		Bucket    string
		AccessKey string
		SecretKey string
		UseSSL    bool
	}

	// Kafka 生产者配置
	KafkaProducer struct {
		Brokers []string
		Topic   string
	}

	// Asynq 配置（异步任务队列）
	Asynq struct {
		Addr     string // Redis 地址
		Password string // Redis 密码
	}

	// Usercenter RPC 客户端配置
	UsercenterRpcConf zrpc.RpcClientConf
}

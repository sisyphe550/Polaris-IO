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

	// File RPC 客户端配置（获取文件信息、下载链接、复制文件）
	FileRpcConf zrpc.RpcClientConf

	// Usercenter RPC 客户端配置（获取用户信息、验证配额）
	UsercenterRpcConf zrpc.RpcClientConf

	// Asynq 配置（异步任务队列）
	Asynq struct {
		Addr     string
		Password string
	}

	// 分享配置
	Share struct {
		BaseUrl    string // 分享链接前缀
		CodeLength int    // 提取码长度
	}
}

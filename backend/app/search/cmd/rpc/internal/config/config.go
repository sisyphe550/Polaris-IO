package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	// Elasticsearch 配置
	ES struct {
		Addresses []string
		Username  string
		Password  string
		Index     string
	}
}

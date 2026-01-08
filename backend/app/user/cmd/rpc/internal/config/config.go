package config

import "github.com/zeromicro/go-zero/zrpc"

type DBConf struct {
	DataSource string
}

type Config struct {
	zrpc.RpcServerConf
	JwtAuth struct {
		AccessSecret string
		AccessExpire int64
	}
	// DB 设为可选：避免开发期不配 DB 就启动失败
	DB *DBConf
}

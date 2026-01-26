package svc

import (
	"polaris-io/backend/app/file/cmd/rpc/fileservice"
	"polaris-io/backend/app/share/cmd/rpc/internal/config"
	"polaris-io/backend/app/share/model"
	"polaris-io/backend/app/user/cmd/rpc/usercenter"
	"polaris-io/backend/pkg/asynqjob"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	// MySQL Model
	ShareModel model.ShareModel

	// File RPC Client（获取文件信息、下载链接、复制文件）
	FileRpc fileservice.FileService

	// Usercenter RPC Client（获取用户信息、验证配额）
	UsercenterRpc usercenter.Usercenter

	// Asynq Client（异步任务队列）
	AsynqClient *asynqjob.AsynqClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化 MySQL 连接
	sqlConn := sqlx.NewMysql(c.DB.DataSource)
	logx.Info("MySQL connected successfully")

	// 初始化 Asynq 客户端
	var asynqClient *asynqjob.AsynqClient
	if c.Asynq.Addr != "" {
		asynqClient = asynqjob.NewAsynqClient(asynqjob.AsynqClientConfig{
			Addr:     c.Asynq.Addr,
			Password: c.Asynq.Password,
		})
		logx.Info("Asynq client initialized successfully")
	}

	return &ServiceContext{
		Config: c,

		// MySQL Model
		ShareModel: model.NewShareModel(sqlConn),

		// File RPC Client
		FileRpc: fileservice.NewFileService(zrpc.MustNewClient(c.FileRpcConf)),

		// Usercenter RPC Client
		UsercenterRpc: usercenter.NewUsercenter(zrpc.MustNewClient(c.UsercenterRpcConf)),

		// Asynq Client
		AsynqClient: asynqClient,
	}
}

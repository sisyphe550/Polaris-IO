package svc

import (
	"context"

	"polaris-io/backend/app/search/cmd/rpc/internal/config"
	"polaris-io/backend/app/search/es"

	"github.com/zeromicro/go-zero/core/logx"
)

type ServiceContext struct {
	Config   config.Config
	ESClient *es.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化 ES 客户端
	esClient, err := es.NewClient(es.Config{
		Addresses: c.ES.Addresses,
		Username:  c.ES.Username,
		Password:  c.ES.Password,
		Index:     c.ES.Index,
	})
	if err != nil {
		logx.Errorf("Failed to create ES client: %v", err)
		panic(err)
	}

	// 确保索引存在
	if err := esClient.CreateIndex(context.Background()); err != nil {
		logx.Errorf("Failed to create ES index: %v", err)
		panic(err)
	}

	logx.Info("ES client initialized and index ready")

	return &ServiceContext{
		Config:   c,
		ESClient: esClient,
	}
}

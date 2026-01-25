package svc

import (
	"polaris-io/backend/app/search/cmd/api/internal/config"
	"polaris-io/backend/app/search/cmd/rpc/searchservice"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config    config.Config
	SearchRpc searchservice.SearchService
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:    c,
		SearchRpc: searchservice.NewSearchService(zrpc.MustNewClient(c.SearchRpcConf)),
	}
}

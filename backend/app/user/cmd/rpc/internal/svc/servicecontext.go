package svc

import (
	"polaris-io/backend/app/user/cmd/rpc/internal/config"
	"polaris-io/backend/app/user/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config

	UserModel      model.UserModel
	UserQuotaModel model.UserQuotaModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlConn := sqlx.NewMysql(c.DB.DataSource)

	return &ServiceContext{
		Config:         c,
		UserModel:      model.NewUserModel(sqlConn),
		UserQuotaModel: model.NewUserQuotaModel(sqlConn),
	}
}

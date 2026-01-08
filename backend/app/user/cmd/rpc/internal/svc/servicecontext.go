package svc

import (
	"shared-board/backend/app/user/cmd/rpc/internal/config"
	"shared-board/backend/app/user/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config     config.Config
	UsersModel model.UsersModel
	MemUsers   *MemUserStore
}

func NewServiceContext(c config.Config) *ServiceContext {
	var usersModel model.UsersModel
	if c.DB != nil && c.DB.DataSource != "" {
		sqlConn := sqlx.NewMysql(c.DB.DataSource)
		usersModel = model.NewUsersModel(sqlConn)
	}

	return &ServiceContext{
		Config:     c,
		UsersModel: usersModel,
		MemUsers:   NewMemUserStore(),
	}
}

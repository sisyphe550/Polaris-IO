package model

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ChatMessagesModel = (*customChatMessagesModel)(nil)

type (
	// ChatMessagesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customChatMessagesModel.
	ChatMessagesModel interface {
		chatMessagesModel
	}

	customChatMessagesModel struct {
		*defaultChatMessagesModel
	}
)

// NewChatMessagesModel returns a model for the database table.
func NewChatMessagesModel(conn sqlx.SqlConn) ChatMessagesModel {
	return &customChatMessagesModel{
		defaultChatMessagesModel: newChatMessagesModel(conn),
	}
}

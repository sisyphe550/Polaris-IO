package model

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ RoomsModel = (*customRoomsModel)(nil)

type (
	// RoomsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customRoomsModel.
	RoomsModel interface {
		roomsModel
	}

	customRoomsModel struct {
		*defaultRoomsModel
	}
)

// NewRoomsModel returns a model for the database table.
func NewRoomsModel(conn sqlx.SqlConn) RoomsModel {
	return &customRoomsModel{
		defaultRoomsModel: newRoomsModel(conn),
	}
}

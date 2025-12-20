package model

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ SnapshotsModel = (*customSnapshotsModel)(nil)

type (
	// SnapshotsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSnapshotsModel.
	SnapshotsModel interface {
		snapshotsModel
	}

	customSnapshotsModel struct {
		*defaultSnapshotsModel
	}
)

// NewSnapshotsModel returns a model for the database table.
func NewSnapshotsModel(conn sqlx.SqlConn) SnapshotsModel {
	return &customSnapshotsModel{
		defaultSnapshotsModel: newSnapshotsModel(conn),
	}
}

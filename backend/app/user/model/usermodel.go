package model

import (
	"context"
	"database/sql"
	"fmt"
	"polaris-io/backend/pkg/globalkey"

	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserModel = (*customUserModel)(nil)

type (
	// UserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserModel.
	UserModel interface {
		userModel // 嵌入自动生成的接口
		// 自定义方法
		FindOneByMobile(ctx context.Context, mobile string) (*User, error)
		InsertWithQuota(ctx context.Context, data *User, quota *UserQuota) (sql.Result, error)
	}

	customUserModel struct {
		*defaultUserModel
	}
)

// NewUserModel returns a model for the database table.
func NewUserModel(conn sqlx.SqlConn) UserModel {
	return &customUserModel{
		defaultUserModel: newUserModel(conn),
	}
}

// FindOneByMobile 根据手机号查询用户（仅查询未删除的）
func (m *customUserModel) FindOneByMobile(ctx context.Context, mobile string) (*User, error) {
	var resp User
	query := fmt.Sprintf("select %s from %s where `mobile` = ? and del_state = ? limit 1", userRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, mobile, globalkey.DelStateNo)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// InsertWithQuota 事务：同时插入用户和配额记录
// 注意：这里需要跨表操作，实际业务中配额表在同一个库
func (m *customUserModel) InsertWithQuota(ctx context.Context, userData *User, quotaData *UserQuota) (sql.Result, error) {
	var result sql.Result
	err := m.Trans(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 插入用户
		res, err := m.Insert(ctx, session, userData)
		if err != nil {
			return err
		}

		// 2. 获取新用户 ID
		userId, err := res.LastInsertId()
		if err != nil {
			return err
		}

		// 3. 插入配额记录
		quotaData.UserId = uint64(userId)
		quotaModel := newUserQuotaModel(m.conn)
		_, err = quotaModel.Insert(ctx, session, quotaData)
		if err != nil {
			return err
		}

		result = res
		return nil
	})

	return result, err
}

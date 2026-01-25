package model

import (
	"context"
	"database/sql"
	"fmt"

	"polaris-io/backend/pkg/globalkey"

	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ShareModel = (*customShareModel)(nil)

type (
	// ShareModel is an interface to be customized, add more methods here,
	// and implement the added methods in customShareModel.
	ShareModel interface {
		shareModel
		// 根据 identity 查询分享
		FindOneByIdentity(ctx context.Context, identity string) (*Share, error)
		// 根据用户 ID 查询分享列表（分页）
		FindListByUserId(ctx context.Context, userId uint64, page, pageSize int64) ([]*Share, int64, error)
		// 批量软删除（取消分享）
		BatchDeleteSoft(ctx context.Context, session sqlx.Session, userId uint64, identities []string) (int64, error)
		// 增加点击次数
		IncrClickNum(ctx context.Context, identity string) error
		// 根据文件 identity 查询分享（检查文件是否已被分享）
		FindOneByRepositoryIdentity(ctx context.Context, userId uint64, repositoryIdentity string) (*Share, error)
	}

	customShareModel struct {
		*defaultShareModel
	}
)

// NewShareModel returns a model for the database table.
func NewShareModel(conn sqlx.SqlConn) ShareModel {
	return &customShareModel{
		defaultShareModel: newShareModel(conn),
	}
}

// FindOneByIdentity 根据 identity 查询分享
func (m *customShareModel) FindOneByIdentity(ctx context.Context, identity string) (*Share, error) {
	query := fmt.Sprintf("select %s from %s where `identity` = ? and del_state = ? limit 1", shareRows, m.table)
	var resp Share
	err := m.conn.QueryRowCtx(ctx, &resp, query, identity, globalkey.DelStateNo)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindListByUserId 根据用户 ID 查询分享列表（分页）
func (m *customShareModel) FindListByUserId(ctx context.Context, userId uint64, page, pageSize int64) ([]*Share, int64, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	// 查询总数
	countQuery := fmt.Sprintf("select count(*) from %s where `user_id` = ? and del_state = ?", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, userId, globalkey.DelStateNo)
	if err != nil {
		return nil, 0, err
	}

	// 查询列表
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and del_state = ? order by id desc limit ? offset ?",
		shareRows, m.table)
	var resp []*Share
	err = m.conn.QueryRowsCtx(ctx, &resp, query, userId, globalkey.DelStateNo, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	return resp, total, nil
}

// BatchDeleteSoft 批量软删除（取消分享）
func (m *customShareModel) BatchDeleteSoft(ctx context.Context, session sqlx.Session, userId uint64, identities []string) (int64, error) {
	if len(identities) == 0 {
		return 0, nil
	}

	// 构建 IN 查询的占位符
	placeholders := ""
	args := make([]interface{}, 0, len(identities)+3)
	args = append(args, globalkey.DelStateYes)
	for i, id := range identities {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args = append(args, id)
	}
	args = append(args, userId, globalkey.DelStateNo)

	query := fmt.Sprintf("update %s set del_state = ?, delete_time = unix_timestamp(), version = version + 1 where identity in (%s) and user_id = ? and del_state = ?",
		m.table, placeholders)

	var result sql.Result
	var err error
	if session != nil {
		result, err = session.ExecCtx(ctx, query, args...)
	} else {
		result, err = m.conn.ExecCtx(ctx, query, args...)
	}
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// IncrClickNum 增加点击次数
func (m *customShareModel) IncrClickNum(ctx context.Context, identity string) error {
	query := fmt.Sprintf("update %s set click_num = click_num + 1 where identity = ? and del_state = ?",
		m.table)
	_, err := m.conn.ExecCtx(ctx, query, identity, globalkey.DelStateNo)
	return err
}

// FindOneByRepositoryIdentity 根据文件 identity 查询分享（检查文件是否已被分享）
func (m *customShareModel) FindOneByRepositoryIdentity(ctx context.Context, userId uint64, repositoryIdentity string) (*Share, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and `repository_identity` = ? and del_state = ? limit 1",
		shareRows, m.table)
	var resp Share
	err := m.conn.QueryRowCtx(ctx, &resp, query, userId, repositoryIdentity, globalkey.DelStateNo)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

package model

import (
	"context"
	"database/sql"
	"fmt"

	"polaris-io/backend/pkg/globalkey"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserRepositoryModel = (*customUserRepositoryModel)(nil)

type (
	// UserRepositoryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserRepositoryModel.
	UserRepositoryModel interface {
		userRepositoryModel
		// 回收站相关方法
		FindTrashList(ctx context.Context, userId uint64, page, pageSize int64) ([]*UserRepository, int64, error)
		FindTrashByIdentity(ctx context.Context, identity string) (*UserRepository, error)
		RestoreFile(ctx context.Context, session sqlx.Session, data *UserRepository) error
		HardDelete(ctx context.Context, session sqlx.Session, id uint64) error
		FindDeletedByIdentities(ctx context.Context, userId uint64, identities []string) ([]*UserRepository, error)
	}

	customUserRepositoryModel struct {
		*defaultUserRepositoryModel
	}
)

// NewUserRepositoryModel returns a model for the database table.
func NewUserRepositoryModel(conn sqlx.SqlConn, c cache.CacheConf) UserRepositoryModel {
	return &customUserRepositoryModel{
		defaultUserRepositoryModel: newUserRepositoryModel(conn, c),
	}
}

// FindTrashList 查询回收站列表（已删除的文件）
func (m *customUserRepositoryModel) FindTrashList(ctx context.Context, userId uint64, page, pageSize int64) ([]*UserRepository, int64, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE user_id = ? AND del_state = ?", m.table)
	var total int64
	err := m.QueryRowNoCacheCtx(ctx, &total, countQuery, userId, globalkey.DelStateYes)
	if err != nil {
		return nil, 0, err
	}

	// 查询列表
	query := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = ? AND del_state = ? ORDER BY delete_time DESC LIMIT ? OFFSET ?",
		userRepositoryRows, m.table)
	var resp []*UserRepository
	err = m.QueryRowsNoCacheCtx(ctx, &resp, query, userId, globalkey.DelStateYes, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	return resp, total, nil
}

// FindTrashByIdentity 根据 identity 查询已删除的文件
func (m *customUserRepositoryModel) FindTrashByIdentity(ctx context.Context, identity string) (*UserRepository, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE identity = ? AND del_state = ? LIMIT 1",
		userRepositoryRows, m.table)
	var resp UserRepository
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, identity, globalkey.DelStateYes)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

// FindDeletedByIdentities 批量查询已删除的文件
func (m *customUserRepositoryModel) FindDeletedByIdentities(ctx context.Context, userId uint64, identities []string) ([]*UserRepository, error) {
	if len(identities) == 0 {
		return []*UserRepository{}, nil
	}

	// 构建 IN 查询
	placeholders := ""
	args := make([]interface{}, 0, len(identities)+2)
	args = append(args, userId, globalkey.DelStateYes)
	for i, id := range identities {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args = append(args, id)
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = ? AND del_state = ? AND identity IN (%s)",
		userRepositoryRows, m.table, placeholders)
	var resp []*UserRepository
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, args...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// RestoreFile 恢复文件（从回收站恢复）
func (m *customUserRepositoryModel) RestoreFile(ctx context.Context, session sqlx.Session, data *UserRepository) error {
	// 清除缓存
	userRepositoryIdKey := fmt.Sprintf("%s%v", cacheUserRepositoryIdPrefix, data.Id)
	userRepositoryIdentityKey := fmt.Sprintf("%s%v", cacheUserRepositoryIdentityPrefix, data.Identity)

	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("UPDATE %s SET del_state = ?, delete_time = ?, version = version + 1 WHERE id = ? AND del_state = ?",
			m.table)
		if session != nil {
			return session.ExecCtx(ctx, query, globalkey.DelStateNo, 0, data.Id, globalkey.DelStateYes)
		}
		return conn.ExecCtx(ctx, query, globalkey.DelStateNo, 0, data.Id, globalkey.DelStateYes)
	}, userRepositoryIdKey, userRepositoryIdentityKey)

	return err
}

// HardDelete 彻底删除文件（物理删除）
func (m *customUserRepositoryModel) HardDelete(ctx context.Context, session sqlx.Session, id uint64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		// 如果 FindOne 找不到（可能已被软删除），直接查询
		query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ? LIMIT 1", userRepositoryRows, m.table)
		err = m.QueryRowNoCacheCtx(ctx, data, query, id)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil // 已经不存在，视为成功
			}
			return err
		}
	}

	// 清除缓存
	userRepositoryIdKey := fmt.Sprintf("%s%v", cacheUserRepositoryIdPrefix, id)
	userRepositoryIdentityKey := fmt.Sprintf("%s%v", cacheUserRepositoryIdentityPrefix, data.Identity)

	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", m.table)
		if session != nil {
			return session.ExecCtx(ctx, query, id)
		}
		return conn.ExecCtx(ctx, query, id)
	}, userRepositoryIdKey, userRepositoryIdentityKey)

	return err
}

package model

import (
	"context"
	"database/sql"
	"fmt"
	"polaris-io/backend/pkg/globalkey"

	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserQuotaModel = (*customUserQuotaModel)(nil)

type (
	// UserQuotaModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserQuotaModel.
	UserQuotaModel interface {
		userQuotaModel
		// 自定义方法
		FindOneByUserId(ctx context.Context, userId uint64) (*UserQuota, error)
		UpdateUsedSize(ctx context.Context, userId uint64, delta int64) error
		DeductQuota(ctx context.Context, userId uint64, size uint64) error
		RefundQuota(ctx context.Context, userId uint64, size uint64) error
	}

	customUserQuotaModel struct {
		*defaultUserQuotaModel
	}
)

// NewUserQuotaModel returns a model for the database table.
func NewUserQuotaModel(conn sqlx.SqlConn) UserQuotaModel {
	return &customUserQuotaModel{
		defaultUserQuotaModel: newUserQuotaModel(conn),
	}
}

// FindOneByUserId 根据用户 ID 查询配额（仅查询未删除的）
func (m *customUserQuotaModel) FindOneByUserId(ctx context.Context, userId uint64) (*UserQuota, error) {
	var resp UserQuota
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and del_state = ? limit 1", userQuotaRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, userId, globalkey.DelStateNo)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// UpdateUsedSize 更新已用容量（带乐观锁）
// delta: 增量，正数表示增加，负数表示减少
func (m *customUserQuotaModel) UpdateUsedSize(ctx context.Context, userId uint64, delta int64) error {
	// 1. 先查询当前记录
	quota, err := m.FindOneByUserId(ctx, userId)
	if err != nil {
		return err
	}

	// 2. 计算新值
	var newUsedSize uint64
	if delta < 0 {
		// 减少容量，防止下溢
		absValue := uint64(-delta)
		if quota.UsedSize < absValue {
			newUsedSize = 0
		} else {
			newUsedSize = quota.UsedSize - absValue
		}
	} else {
		newUsedSize = quota.UsedSize + uint64(delta)
	}

	// 3. 使用乐观锁更新
	quota.UsedSize = newUsedSize
	return m.UpdateWithVersion(ctx, nil, quota)
}

// DeductQuota 扣减配额（上传文件时调用）
// 使用原子操作防止超额
func (m *customUserQuotaModel) DeductQuota(ctx context.Context, userId uint64, size uint64) error {
	// 使用原子 SQL 语句，防止并发超额
	query := fmt.Sprintf(`
		UPDATE %s 
		SET used_size = used_size + ?, version = version + 1 
		WHERE user_id = ? 
		  AND del_state = ? 
		  AND used_size + ? <= total_size
	`, m.table)

	result, err := m.conn.ExecCtx(ctx, query, size, userId, globalkey.DelStateNo, size)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrQuotaExceeded
	}

	return nil
}

// RefundQuota 退还配额（删除文件时调用）
func (m *customUserQuotaModel) RefundQuota(ctx context.Context, userId uint64, size uint64) error {
	query := fmt.Sprintf(`
		UPDATE %s 
		SET used_size = CASE 
			WHEN used_size >= ? THEN used_size - ? 
			ELSE 0 
		END,
		version = version + 1 
		WHERE user_id = ? AND del_state = ?
	`, m.table)

	result, err := m.conn.ExecCtx(ctx, query, size, size, userId, globalkey.DelStateNo)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

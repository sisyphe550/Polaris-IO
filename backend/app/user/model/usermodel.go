package model

import (
	"context"
	"database/sql"
	"fmt"
	"polaris-io/backend/pkg/globalkey"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserModel = (*customUserModel)(nil)

// 缓存 key 前缀
const (
	cacheUserIdPrefix     = "cache:user:id:"
	cacheUserMobilePrefix = "cache:user:mobile:"
)

type (
	// UserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserModel.
	UserModel interface {
		userModel // 嵌入自动生成的接口
		// 自定义方法
		FindOneByMobile(ctx context.Context, mobile string) (*User, error)
		InsertWithQuota(ctx context.Context, data *User, quota *UserQuota) (sql.Result, error)
		// 带缓存的查询方法
		FindOneWithCache(ctx context.Context, id uint64) (*User, error)
		FindOneByMobileWithCache(ctx context.Context, mobile string) (*User, error)
		// 缓存管理
		InvalidateCache(ctx context.Context, id uint64, mobile string) error
	}

	customUserModel struct {
		*defaultUserModel
		cachedConn sqlc.CachedConn // 带缓存的连接
	}
)

// NewUserModel returns a model for the database table.
// cacheConf 可以为空，此时不启用缓存
func NewUserModel(conn sqlx.SqlConn, cacheConf ...cache.CacheConf) UserModel {
	m := &customUserModel{
		defaultUserModel: newUserModel(conn),
	}

	// 如果提供了缓存配置，则初始化 CachedConn
	if len(cacheConf) > 0 && len(cacheConf[0]) > 0 {
		m.cachedConn = sqlc.NewConn(conn, cacheConf[0])
	}

	return m
}

// FindOneByMobile 根据手机号查询用户（仅查询未删除的，不走缓存）
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

// FindOneWithCache 根据 ID 查询用户（走缓存）
func (m *customUserModel) FindOneWithCache(ctx context.Context, id uint64) (*User, error) {
	// 如果缓存未启用，直接查数据库
	if m.cachedConn == (sqlc.CachedConn{}) {
		return m.FindOne(ctx, id)
	}

	cacheKey := fmt.Sprintf("%s%d", cacheUserIdPrefix, id)
	var resp User

	err := m.cachedConn.QueryRowCtx(ctx, &resp, cacheKey, func(ctx context.Context, conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? and del_state = ? limit 1", userRows, m.table)
		return conn.QueryRowCtx(ctx, v, query, id, globalkey.DelStateNo)
	})

	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindOneByMobileWithCache 根据手机号查询用户（走缓存）
func (m *customUserModel) FindOneByMobileWithCache(ctx context.Context, mobile string) (*User, error) {
	// 如果缓存未启用，直接查数据库
	if m.cachedConn == (sqlc.CachedConn{}) {
		return m.FindOneByMobile(ctx, mobile)
	}

	cacheKey := fmt.Sprintf("%s%s", cacheUserMobilePrefix, mobile)
	var resp User

	err := m.cachedConn.QueryRowCtx(ctx, &resp, cacheKey, func(ctx context.Context, conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where `mobile` = ? and del_state = ? limit 1", userRows, m.table)
		return conn.QueryRowCtx(ctx, v, query, mobile, globalkey.DelStateNo)
	})

	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// InvalidateCache 使缓存失效
func (m *customUserModel) InvalidateCache(ctx context.Context, id uint64, mobile string) error {
	if m.cachedConn == (sqlc.CachedConn{}) {
		return nil
	}

	keys := make([]string, 0, 2)
	if id > 0 {
		keys = append(keys, fmt.Sprintf("%s%d", cacheUserIdPrefix, id))
	}
	if mobile != "" {
		keys = append(keys, fmt.Sprintf("%s%s", cacheUserMobilePrefix, mobile))
	}

	return m.cachedConn.DelCacheCtx(ctx, keys...)
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

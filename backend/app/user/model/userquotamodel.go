package model

import (
	"context"
	"database/sql"
	"fmt"
	"polaris-io/backend/pkg/globalkey"
	"polaris-io/backend/pkg/quotacache"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
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
		FindOneByUserIdWithCache(ctx context.Context, userId uint64) (*UserQuota, error)
		UpdateUsedSize(ctx context.Context, userId uint64, delta int64) error
		DeductQuota(ctx context.Context, userId uint64, size uint64) error
		RefundQuota(ctx context.Context, userId uint64, size uint64) error
		// 缓存相关
		InvalidateCache(ctx context.Context, userId uint64) error
		WarmUpCache(ctx context.Context, userId uint64) error
	}

	customUserQuotaModel struct {
		*defaultUserQuotaModel
		quotaCache *quotacache.QuotaCache
	}
)

// NewUserQuotaModel returns a model for the database table.
// redisClient 可以为 nil，此时不启用缓存
func NewUserQuotaModel(conn sqlx.SqlConn, redisClient *redis.Redis) UserQuotaModel {
	var qc *quotacache.QuotaCache
	if redisClient != nil {
		qc = quotacache.NewQuotaCache(redisClient)
	}
	return &customUserQuotaModel{
		defaultUserQuotaModel: newUserQuotaModel(conn),
		quotaCache:            qc,
	}
}

// FindOneByUserId 根据用户 ID 查询配额（仅查询未删除的，不走缓存）
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

// FindOneByUserIdWithCache 根据用户 ID 查询配额（优先走缓存）
func (m *customUserQuotaModel) FindOneByUserIdWithCache(ctx context.Context, userId uint64) (*UserQuota, error) {
	// 如果缓存未启用，直接查数据库
	if m.quotaCache == nil {
		return m.FindOneByUserId(ctx, userId)
	}

	// 1. 先查缓存
	cached, exists, err := m.quotaCache.Get(ctx, userId)
	if err != nil {
		logx.WithContext(ctx).Errorf("FindOneByUserIdWithCache: cache get error: %v", err)
		// 缓存出错，降级查数据库
		return m.FindOneByUserId(ctx, userId)
	}

	if exists {
		// 缓存命中，构造返回对象
		// 注意：缓存只存储了 TotalSize 和 UsedSize，其他字段需要查数据库
		// 但对于配额查询场景，这两个字段已经足够
		logx.WithContext(ctx).Debugf("FindOneByUserIdWithCache: cache hit for userId=%d", userId)
		return &UserQuota{
			UserId:    userId,
			TotalSize: cached.TotalSize,
			UsedSize:  cached.UsedSize,
		}, nil
	}

	// 2. 缓存未命中，查数据库
	logx.WithContext(ctx).Debugf("FindOneByUserIdWithCache: cache miss for userId=%d", userId)
	quota, err := m.FindOneByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	// 3. 写入缓存
	cacheErr := m.quotaCache.Set(ctx, userId, &quotacache.QuotaInfo{
		TotalSize: quota.TotalSize,
		UsedSize:  quota.UsedSize,
	})
	if cacheErr != nil {
		logx.WithContext(ctx).Errorf("FindOneByUserIdWithCache: cache set error: %v", cacheErr)
		// 缓存写入失败不影响返回
	}

	return quota, nil
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
// 使用 Redis 缓存 + 数据库原子操作，实现高性能配额扣减
func (m *customUserQuotaModel) DeductQuota(ctx context.Context, userId uint64, size uint64) error {
	// 1. 如果缓存启用，先尝试在缓存中扣减
	if m.quotaCache != nil {
		result, err := m.quotaCache.DeductQuota(ctx, userId, size)
		if err != nil {
			logx.WithContext(ctx).Errorf("DeductQuota: cache deduct error: %v", err)
			// 缓存操作失败，降级到数据库
		} else {
			switch result {
			case quotacache.DeductSuccess:
				// 缓存扣减成功，异步更新数据库（或者同步更新确保一致性）
				// 这里选择同步更新数据库，确保数据一致性
				dbErr := m.deductQuotaDB(ctx, userId, size)
				if dbErr != nil {
					// 数据库更新失败，需要回滚缓存
					logx.WithContext(ctx).Errorf("DeductQuota: db deduct error, rolling back cache: %v", dbErr)
					_ = m.quotaCache.RefundQuota(ctx, userId, size)
					return dbErr
				}
				return nil

			case quotacache.DeductQuotaExceeded:
				// 缓存显示配额不足
				return ErrQuotaExceeded

			case quotacache.DeductCacheMiss:
				// 缓存不存在，需要预热缓存后重试
				logx.WithContext(ctx).Debugf("DeductQuota: cache miss, warming up cache for userId=%d", userId)
				_ = m.WarmUpCache(ctx, userId)
				// 预热后重试一次缓存扣减
				retryResult, retryErr := m.quotaCache.DeductQuota(ctx, userId, size)
				if retryErr == nil && retryResult == quotacache.DeductSuccess {
					dbErr := m.deductQuotaDB(ctx, userId, size)
					if dbErr != nil {
						_ = m.quotaCache.RefundQuota(ctx, userId, size)
						return dbErr
					}
					return nil
				} else if retryResult == quotacache.DeductQuotaExceeded {
					return ErrQuotaExceeded
				}
				// 重试也失败，降级到纯数据库操作
			}
		}
	}

	// 2. 降级：直接使用数据库原子操作
	return m.deductQuotaDB(ctx, userId, size)
}

// deductQuotaDB 数据库层面的配额扣减（原子操作）
func (m *customUserQuotaModel) deductQuotaDB(ctx context.Context, userId uint64, size uint64) error {
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
// 先更新数据库，再更新缓存
func (m *customUserQuotaModel) RefundQuota(ctx context.Context, userId uint64, size uint64) error {
	// 1. 先更新数据库（数据库是最终一致性的保证）
	err := m.refundQuotaDB(ctx, userId, size)
	if err != nil {
		return err
	}

	// 2. 更新缓存（失败不影响主流程）
	if m.quotaCache != nil {
		cacheErr := m.quotaCache.RefundQuota(ctx, userId, size)
		if cacheErr != nil {
			logx.WithContext(ctx).Errorf("RefundQuota: cache refund error: %v", cacheErr)
			// 缓存更新失败，删除缓存让下次重新加载
			_ = m.quotaCache.Delete(ctx, userId)
		}
	}

	return nil
}

// refundQuotaDB 数据库层面的配额退还
func (m *customUserQuotaModel) refundQuotaDB(ctx context.Context, userId uint64, size uint64) error {
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

// InvalidateCache 使缓存失效
func (m *customUserQuotaModel) InvalidateCache(ctx context.Context, userId uint64) error {
	if m.quotaCache == nil {
		return nil
	}
	return m.quotaCache.Delete(ctx, userId)
}

// WarmUpCache 预热缓存（从数据库加载到缓存）
func (m *customUserQuotaModel) WarmUpCache(ctx context.Context, userId uint64) error {
	if m.quotaCache == nil {
		return nil
	}

	// 从数据库查询
	quota, err := m.FindOneByUserId(ctx, userId)
	if err != nil {
		return err
	}

	// 写入缓存
	return m.quotaCache.Set(ctx, userId, &quotacache.QuotaInfo{
		TotalSize: quota.TotalSize,
		UsedSize:  quota.UsedSize,
	})
}

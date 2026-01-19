package quotacache

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	// QuotaCacheKeyPrefix Redis 配额缓存 key 前缀
	QuotaCacheKeyPrefix = "user:quota:"

	// 字段名
	FieldTotalSize = "total"
	FieldUsedSize  = "used"

	// 缓存过期时间 (秒) - 24小时
	QuotaCacheExpire = 86400
)

// QuotaCache 配额缓存
type QuotaCache struct {
	redisClient *redis.Redis
}

// QuotaInfo 配额信息
type QuotaInfo struct {
	TotalSize uint64
	UsedSize  uint64
}

// NewQuotaCache 创建配额缓存实例
func NewQuotaCache(redisClient *redis.Redis) *QuotaCache {
	return &QuotaCache{
		redisClient: redisClient,
	}
}

// GetCacheKey 获取缓存 key
func GetCacheKey(userId uint64) string {
	return fmt.Sprintf("%s%d", QuotaCacheKeyPrefix, userId)
}

// Get 从缓存获取配额
// 返回值: quota, exists, error
func (c *QuotaCache) Get(ctx context.Context, userId uint64) (*QuotaInfo, bool, error) {
	key := GetCacheKey(userId)

	// 使用 HGETALL 获取所有字段
	result, err := c.redisClient.HgetallCtx(ctx, key)
	if err != nil {
		return nil, false, err
	}

	// 缓存不存在
	if len(result) == 0 {
		return nil, false, nil
	}

	// 解析缓存数据
	totalStr, ok1 := result[FieldTotalSize]
	usedStr, ok2 := result[FieldUsedSize]
	if !ok1 || !ok2 {
		// 缓存数据不完整，删除并返回不存在
		_, _ = c.redisClient.DelCtx(ctx, key)
		return nil, false, nil
	}

	total, err := strconv.ParseUint(totalStr, 10, 64)
	if err != nil {
		_, _ = c.redisClient.DelCtx(ctx, key)
		return nil, false, nil
	}

	used, err := strconv.ParseUint(usedStr, 10, 64)
	if err != nil {
		_, _ = c.redisClient.DelCtx(ctx, key)
		return nil, false, nil
	}

	return &QuotaInfo{
		TotalSize: total,
		UsedSize:  used,
	}, true, nil
}

// Set 设置缓存
func (c *QuotaCache) Set(ctx context.Context, userId uint64, quota *QuotaInfo) error {
	key := GetCacheKey(userId)

	// 使用 HSET 设置多个字段
	err := c.redisClient.HsetCtx(ctx, key, FieldTotalSize, strconv.FormatUint(quota.TotalSize, 10))
	if err != nil {
		return err
	}
	err = c.redisClient.HsetCtx(ctx, key, FieldUsedSize, strconv.FormatUint(quota.UsedSize, 10))
	if err != nil {
		return err
	}

	// 设置过期时间
	return c.redisClient.ExpireCtx(ctx, key, QuotaCacheExpire)
}

// Delete 删除缓存
func (c *QuotaCache) Delete(ctx context.Context, userId uint64) error {
	key := GetCacheKey(userId)
	_, err := c.redisClient.DelCtx(ctx, key)
	return err
}

// DeductQuotaScript Lua 脚本: 扣减配额 (原子操作)
// KEYS[1]: 缓存 key
// ARGV[1]: 要扣减的大小
// 返回值: 0=成功, -1=配额不足, -2=缓存不存在
const DeductQuotaScript = `
local key = KEYS[1]
local size = tonumber(ARGV[1])

-- 检查缓存是否存在
local exists = redis.call('EXISTS', key)
if exists == 0 then
    return -2
end

-- 获取当前值
local used = tonumber(redis.call('HGET', key, 'used') or 0)
local total = tonumber(redis.call('HGET', key, 'total') or 0)

-- 检查配额
if used + size > total then
    return -1
end

-- 扣减配额
redis.call('HINCRBY', key, 'used', size)
redis.call('EXPIRE', key, 86400)

return 0
`

// RefundQuotaScript Lua 脚本: 退还配额 (原子操作)
// KEYS[1]: 缓存 key
// ARGV[1]: 要退还的大小
// 返回值: 0=成功, -2=缓存不存在
const RefundQuotaScript = `
local key = KEYS[1]
local size = tonumber(ARGV[1])

-- 检查缓存是否存在
local exists = redis.call('EXISTS', key)
if exists == 0 then
    return -2
end

-- 获取当前已用空间
local used = tonumber(redis.call('HGET', key, 'used') or 0)

-- 计算新值，防止下溢
local newUsed = used - size
if newUsed < 0 then
    newUsed = 0
end

-- 更新已用空间
redis.call('HSET', key, 'used', newUsed)
redis.call('EXPIRE', key, 86400)

return 0
`

// DeductQuotaResult 扣减配额的结果
type DeductQuotaResult int

const (
	DeductSuccess       DeductQuotaResult = 0  // 成功
	DeductQuotaExceeded DeductQuotaResult = -1 // 配额不足
	DeductCacheMiss     DeductQuotaResult = -2 // 缓存不存在
)

// DeductQuota 从缓存扣减配额 (原子操作)
// 返回: DeductQuotaResult, error
func (c *QuotaCache) DeductQuota(ctx context.Context, userId uint64, size uint64) (DeductQuotaResult, error) {
	key := GetCacheKey(userId)

	result, err := c.redisClient.EvalCtx(ctx, DeductQuotaScript, []string{key}, size)
	if err != nil {
		return DeductCacheMiss, err
	}

	code, ok := result.(int64)
	if !ok {
		logx.WithContext(ctx).Errorf("DeductQuota: unexpected result type: %T", result)
		return DeductCacheMiss, nil
	}

	return DeductQuotaResult(code), nil
}

// RefundQuota 退还配额到缓存 (原子操作)
func (c *QuotaCache) RefundQuota(ctx context.Context, userId uint64, size uint64) error {
	key := GetCacheKey(userId)

	result, err := c.redisClient.EvalCtx(ctx, RefundQuotaScript, []string{key}, size)
	if err != nil {
		return err
	}

	code, ok := result.(int64)
	if !ok {
		logx.WithContext(ctx).Errorf("RefundQuota: unexpected result type: %T", result)
		// 缓存不存在时不报错，因为退还不是强依赖缓存
		return nil
	}

	if code == -2 {
		// 缓存不存在，不报错
		logx.WithContext(ctx).Infof("RefundQuota: cache miss for userId=%d", userId)
	}

	return nil
}

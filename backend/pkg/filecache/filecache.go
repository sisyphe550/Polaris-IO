package filecache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	// 秒传缓存 key 前缀
	FileMetaCacheKeyPrefix = "file:meta:"
	// 文件列表缓存 key 前缀
	FileListCacheKeyPrefix = "file:list:"

	// 秒传缓存过期时间（24小时）
	FileMetaCacheExpire = 86400
	// 文件列表缓存过期时间（5分钟）
	FileListCacheExpire = 300
)

// FileCache 文件缓存
type FileCache struct {
	redisClient *redis.Redis
}

// FileMetaCache 秒传缓存数据结构
type FileMetaCache struct {
	ID       string `json:"id"`
	Hash     string `json:"hash"`
	Size     uint64 `json:"size"`
	S3Key    string `json:"s3Key"`
	Ext      string `json:"ext"`
	MimeType string `json:"mimeType"`
	RefCount int64  `json:"refCount"`
}

// FileListCache 文件列表缓存数据结构
type FileListCache struct {
	List  []FileItemCache `json:"list"`
	Total int64           `json:"total"`
}

// FileItemCache 文件项缓存
type FileItemCache struct {
	Id         int64  `json:"id"`
	Identity   string `json:"identity"`
	Hash       string `json:"hash"`
	UserId     int64  `json:"userId"`
	ParentId   int64  `json:"parentId"`
	Name       string `json:"name"`
	Ext        string `json:"ext"`
	Size       uint64 `json:"size"`
	Path       string `json:"path"`
	IsDir      bool   `json:"isDir"`
	CreateTime int64  `json:"createTime"`
	UpdateTime int64  `json:"updateTime"`
}

// NewFileCache 创建文件缓存实例
func NewFileCache(redisClient *redis.Redis) *FileCache {
	return &FileCache{
		redisClient: redisClient,
	}
}

// ==================== 秒传缓存 ====================

// GetFileMetaCacheKey 获取秒传缓存 key
func GetFileMetaCacheKey(hash string) string {
	return fmt.Sprintf("%s%s", FileMetaCacheKeyPrefix, hash)
}

// GetFileMeta 从缓存获取文件元数据
func (c *FileCache) GetFileMeta(ctx context.Context, hash string) (*FileMetaCache, bool, error) {
	key := GetFileMetaCacheKey(hash)

	data, err := c.redisClient.GetCtx(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}
		return nil, false, err
	}

	if data == "" {
		return nil, false, nil
	}

	var meta FileMetaCache
	if err := json.Unmarshal([]byte(data), &meta); err != nil {
		// 缓存数据损坏，删除
		_, _ = c.redisClient.DelCtx(ctx, key)
		return nil, false, nil
	}

	return &meta, true, nil
}

// SetFileMeta 设置秒传缓存
func (c *FileCache) SetFileMeta(ctx context.Context, hash string, meta *FileMetaCache) error {
	key := GetFileMetaCacheKey(hash)

	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	return c.redisClient.SetexCtx(ctx, key, string(data), FileMetaCacheExpire)
}

// DeleteFileMeta 删除秒传缓存
func (c *FileCache) DeleteFileMeta(ctx context.Context, hash string) error {
	key := GetFileMetaCacheKey(hash)
	_, err := c.redisClient.DelCtx(ctx, key)
	return err
}

// ==================== 文件列表缓存 ====================

// GetFileListCacheKey 获取文件列表缓存 key
// 包含用户ID、父目录ID、页码、每页数量、排序
func GetFileListCacheKey(userId, parentId, page, pageSize int64, orderBy string) string {
	return fmt.Sprintf("%s%d:%d:%d:%d:%s", FileListCacheKeyPrefix, userId, parentId, page, pageSize, orderBy)
}

// GetFileList 从缓存获取文件列表
func (c *FileCache) GetFileList(ctx context.Context, userId, parentId, page, pageSize int64, orderBy string) (*FileListCache, bool, error) {
	key := GetFileListCacheKey(userId, parentId, page, pageSize, orderBy)

	data, err := c.redisClient.GetCtx(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}
		return nil, false, err
	}

	if data == "" {
		return nil, false, nil
	}

	var list FileListCache
	if err := json.Unmarshal([]byte(data), &list); err != nil {
		// 缓存数据损坏，删除
		_, _ = c.redisClient.DelCtx(ctx, key)
		return nil, false, nil
	}

	return &list, true, nil
}

// SetFileList 设置文件列表缓存
func (c *FileCache) SetFileList(ctx context.Context, userId, parentId, page, pageSize int64, orderBy string, list *FileListCache) error {
	key := GetFileListCacheKey(userId, parentId, page, pageSize, orderBy)

	data, err := json.Marshal(list)
	if err != nil {
		return err
	}

	return c.redisClient.SetexCtx(ctx, key, string(data), FileListCacheExpire)
}

// InvalidateUserFileListCache 清除用户某个目录下的所有文件列表缓存
// 使用通配符删除: file:list:{userId}:{parentId}:*
func (c *FileCache) InvalidateUserFileListCache(ctx context.Context, userId, parentId int64) error {
	pattern := fmt.Sprintf("%s%d:%d:*", FileListCacheKeyPrefix, userId, parentId)

	keys, err := c.redisClient.KeysCtx(ctx, pattern)
	if err != nil {
		logx.WithContext(ctx).Errorf("InvalidateUserFileListCache Keys error: %v", err)
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	// 批量删除
	_, err = c.redisClient.DelCtx(ctx, keys...)
	if err != nil {
		logx.WithContext(ctx).Errorf("InvalidateUserFileListCache Del error: %v", err)
		return err
	}

	logx.WithContext(ctx).Debugf("InvalidateUserFileListCache: deleted %d keys for userId=%d, parentId=%d", len(keys), userId, parentId)
	return nil
}

// InvalidateAllUserFileListCache 清除用户所有文件列表缓存
// 用于移动文件等跨目录操作
func (c *FileCache) InvalidateAllUserFileListCache(ctx context.Context, userId int64) error {
	pattern := fmt.Sprintf("%s%d:*", FileListCacheKeyPrefix, userId)

	keys, err := c.redisClient.KeysCtx(ctx, pattern)
	if err != nil {
		logx.WithContext(ctx).Errorf("InvalidateAllUserFileListCache Keys error: %v", err)
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	// 批量删除
	_, err = c.redisClient.DelCtx(ctx, keys...)
	if err != nil {
		logx.WithContext(ctx).Errorf("InvalidateAllUserFileListCache Del error: %v", err)
		return err
	}

	logx.WithContext(ctx).Debugf("InvalidateAllUserFileListCache: deleted %d keys for userId=%d", len(keys), userId)
	return nil
}

// ==================== 辅助方法 ====================

// SetWithExpire 通用缓存设置（带过期时间）
func (c *FileCache) SetWithExpire(ctx context.Context, key string, value interface{}, expireSeconds int) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.redisClient.SetexCtx(ctx, key, string(data), expireSeconds)
}

// Get 通用缓存获取
func (c *FileCache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	data, err := c.redisClient.GetCtx(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	if data == "" {
		return false, nil
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return false, err
	}

	return true, nil
}

// Delete 通用缓存删除
func (c *FileCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	_, err := c.redisClient.DelCtx(ctx, keys...)
	return err
}

// TTL 获取缓存剩余时间
func (c *FileCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.redisClient.TtlCtx(ctx, key)
	if err != nil {
		return 0, err
	}
	return time.Duration(ttl) * time.Second, nil
}

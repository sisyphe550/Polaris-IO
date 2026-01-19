package svc

import (
	"polaris-io/backend/app/user/cmd/rpc/internal/config"
	"polaris-io/backend/app/user/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config

	UserModel      model.UserModel
	UserQuotaModel model.UserQuotaModel

	// Redis 客户端（用于配额缓存）
	RedisClient *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlConn := sqlx.NewMysql(c.DB.DataSource)

	// 初始化 Redis 客户端（从 Cache 配置中获取，与 sqlc 共用同一个 Redis）
	var redisClient *redis.Redis
	if len(c.Cache) > 0 && c.Cache[0].Host != "" {
		// 从 CacheConf 中提取 Redis 配置
		redisConf := redis.RedisConf{
			Host: c.Cache[0].Host,
			Type: "node",
			Pass: c.Cache[0].Pass,
		}
		redisClient = redis.MustNewRedis(redisConf)
		logx.Info("Redis client initialized successfully for quota cache")
	}

	return &ServiceContext{
		Config: c,

		// UserModel 使用 sqlc 缓存
		UserModel: model.NewUserModel(sqlConn, c.Cache),

		// UserQuotaModel 使用 Redis 配额缓存
		UserQuotaModel: model.NewUserQuotaModel(sqlConn, redisClient),

		RedisClient: redisClient,
	}
}

package svc

import (
	"context"

	"polaris-io/backend/app/file/cmd/rpc/internal/config"
	"polaris-io/backend/app/file/model"
	fileMongo "polaris-io/backend/app/file/mongo"
	"polaris-io/backend/app/user/cmd/rpc/usercenter"
	"polaris-io/backend/pkg/asynqjob"
	"polaris-io/backend/pkg/filecache"
	"polaris-io/backend/pkg/kafka"
	"polaris-io/backend/pkg/s3client"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ServiceContext struct {
	Config config.Config

	// MySQL Model
	UserRepositoryModel model.UserRepositoryModel

	// MongoDB
	MongoClient   *mongo.Client
	FileMetaModel fileMongo.FileMetaModel

	// S3 Client
	S3Client *s3client.S3Client

	// Kafka Producer
	KafkaProducer *kafka.Producer

	// Asynq Client（异步任务队列）
	AsynqClient *asynqjob.AsynqClient

	// Usercenter RPC Client
	UsercenterRpc usercenter.Usercenter

	// Redis 客户端
	RedisClient *redis.Redis

	// 文件缓存
	FileCache *filecache.FileCache
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化 MySQL 连接
	sqlConn := sqlx.NewMysql(c.DB.DataSource)

	// 初始化 MongoDB 连接
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(c.MongoDB.Uri))
	if err != nil {
		logx.Errorf("Failed to connect to MongoDB: %v", err)
		panic(err)
	}
	// Ping MongoDB 确保连接成功
	if err := mongoClient.Ping(context.Background(), nil); err != nil {
		logx.Errorf("Failed to ping MongoDB: %v", err)
		panic(err)
	}
	logx.Info("MongoDB connected successfully")

	// 获取数据库实例
	mongoDatabase := mongoClient.Database(c.MongoDB.Database)

	// 确保 MongoDB 索引存在
	if err := fileMongo.EnsureIndexes(context.Background(), mongoDatabase); err != nil {
		logx.Errorf("Failed to ensure MongoDB indexes: %v", err)
		// 不 panic，索引创建失败不影响启动
	}

	// 初始化 S3 Client
	s3Client, err := s3client.NewS3Client(s3client.S3Config{
		Endpoint:  c.S3.Endpoint,
		Region:    c.S3.Region,
		Bucket:    c.S3.Bucket,
		AccessKey: c.S3.AccessKey,
		SecretKey: c.S3.SecretKey,
		UseSSL:    c.S3.UseSSL,
	})
	if err != nil {
		logx.Errorf("Failed to create S3 client: %v", err)
		panic(err)
	}
	logx.Info("S3 client initialized successfully")

	// 初始化 Kafka Producer
	kafkaProducer := kafka.NewProducer(kafka.ProducerConfig{
		Brokers: c.KafkaProducer.Brokers,
		Topic:   c.KafkaProducer.Topic,
	})
	logx.Info("Kafka producer initialized successfully")

	// 初始化 Redis 客户端（从 Cache 配置中获取）
	var redisClient *redis.Redis
	var fc *filecache.FileCache
	if len(c.Cache) > 0 && c.Cache[0].Host != "" {
		redisConf := redis.RedisConf{
			Host: c.Cache[0].Host,
			Type: "node",
			Pass: c.Cache[0].Pass,
		}
		redisClient = redis.MustNewRedis(redisConf)
		fc = filecache.NewFileCache(redisClient)
		logx.Info("Redis client and FileCache initialized successfully")
	}

	// 初始化 Asynq 客户端
	var asynqClient *asynqjob.AsynqClient
	if c.Asynq.Addr != "" {
		asynqClient = asynqjob.NewAsynqClient(asynqjob.AsynqClientConfig{
			Addr:     c.Asynq.Addr,
			Password: c.Asynq.Password,
		})
		logx.Info("Asynq client initialized successfully")
	}

	return &ServiceContext{
		Config: c,

		// MySQL Model
		UserRepositoryModel: model.NewUserRepositoryModel(sqlConn, c.Cache),

		// MongoDB
		MongoClient:   mongoClient,
		FileMetaModel: fileMongo.NewFileMetaModel(mongoDatabase),

		// S3 Client
		S3Client: s3Client,

		// Kafka Producer
		KafkaProducer: kafkaProducer,

		// Asynq Client
		AsynqClient: asynqClient,

		// Usercenter RPC Client
		UsercenterRpc: usercenter.NewUsercenter(zrpc.MustNewClient(c.UsercenterRpcConf)),

		// Redis 客户端和缓存
		RedisClient: redisClient,
		FileCache:   fc,
	}
}

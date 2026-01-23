package svc

import (
	"context"

	"polaris-io/backend/app/file/model"
	fileMongo "polaris-io/backend/app/file/mongo"
	"polaris-io/backend/app/mqueue/cmd/job/internal/config"
	"polaris-io/backend/app/user/cmd/rpc/usercenter"
	"polaris-io/backend/pkg/s3client"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ServiceContext struct {
	Config config.Config

	// S3 Client
	S3Client *s3client.S3Client

	// MongoDB
	MongoClient   *mongo.Client
	FileMetaModel fileMongo.FileMetaModel

	// MySQL Model
	UserRepositoryModel model.UserRepositoryModel

	// Usercenter RPC Client
	UsercenterRpc usercenter.Usercenter
}

func NewServiceContext(c config.Config) *ServiceContext {
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

	// 初始化 MongoDB 连接
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(c.MongoDB.Uri))
	if err != nil {
		logx.Errorf("Failed to connect to MongoDB: %v", err)
		panic(err)
	}
	if err := mongoClient.Ping(context.Background(), nil); err != nil {
		logx.Errorf("Failed to ping MongoDB: %v", err)
		panic(err)
	}
	logx.Info("MongoDB connected successfully")
	mongoDatabase := mongoClient.Database(c.MongoDB.Database)

	// 初始化 MySQL 连接
	sqlConn := sqlx.NewMysql(c.DB.DataSource)

	return &ServiceContext{
		Config: c,

		// S3 Client
		S3Client: s3Client,

		// MongoDB
		MongoClient:   mongoClient,
		FileMetaModel: fileMongo.NewFileMetaModel(mongoDatabase),

		// MySQL Model
		UserRepositoryModel: model.NewUserRepositoryModel(sqlConn, c.Cache),

		// Usercenter RPC Client
		UsercenterRpc: usercenter.NewUsercenter(zrpc.MustNewClient(c.UsercenterRpcConf)),
	}
}

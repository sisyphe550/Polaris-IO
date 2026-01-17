package s3client

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Config S3 配置
type S3Config struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

// S3Client S3 客户端封装
type S3Client struct {
	client *s3.Client
	config S3Config
}

// NewS3Client 创建 S3 客户端
func NewS3Client(cfg S3Config) (*S3Client, error) {
	// 构建 endpoint URL
	scheme := "http"
	if cfg.UseSSL {
		scheme = "https"
	}
	endpointURL := fmt.Sprintf("%s://%s", scheme, cfg.Endpoint)

	// 创建自定义 endpoint resolver
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               endpointURL,
			HostnameImmutable: true,
			SigningRegion:     cfg.Region,
		}, nil
	})

	// 加载配置
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKey,
			cfg.SecretKey,
			"",
		)),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// 创建 S3 客户端
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true // 使用路径风格访问 (对 MinIO/Garage 等兼容存储很重要)
	})

	return &S3Client{
		client: client,
		config: cfg,
	}, nil
}

// GetPresignedUploadURL 获取预签名上传 URL
func (c *S3Client) GetPresignedUploadURL(ctx context.Context, key string, contentType string, expireSeconds int64) (string, error) {
	presignClient := s3.NewPresignClient(c.client)

	input := &s3.PutObjectInput{
		Bucket: aws.String(c.config.Bucket),
		Key:    aws.String(key),
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	presignedReq, err := presignClient.PresignPutObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expireSeconds) * time.Second
	})
	if err != nil {
		return "", fmt.Errorf("failed to presign put object: %w", err)
	}

	return presignedReq.URL, nil
}

// GetPresignedDownloadURL 获取预签名下载 URL
func (c *S3Client) GetPresignedDownloadURL(ctx context.Context, key string, expireSeconds int64) (string, error) {
	presignClient := s3.NewPresignClient(c.client)

	input := &s3.GetObjectInput{
		Bucket: aws.String(c.config.Bucket),
		Key:    aws.String(key),
	}

	presignedReq, err := presignClient.PresignGetObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expireSeconds) * time.Second
	})
	if err != nil {
		return "", fmt.Errorf("failed to presign get object: %w", err)
	}

	return presignedReq.URL, nil
}

// PutObject 上传对象 (用于小文件直接上传)
func (c *S3Client) PutObject(ctx context.Context, key string, body io.Reader, contentType string) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(c.config.Bucket),
		Key:    aws.String(key),
		Body:   body,
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	_, err := c.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put object: %w", err)
	}
	return nil
}

// DeleteObject 删除对象
func (c *S3Client) DeleteObject(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(c.config.Bucket),
		Key:    aws.String(key),
	}

	_, err := c.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// HeadObject 检查对象是否存在
func (c *S3Client) HeadObject(ctx context.Context, key string) (bool, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(c.config.Bucket),
		Key:    aws.String(key),
	}

	_, err := c.client.HeadObject(ctx, input)
	if err != nil {
		// TODO: 检查是否是 NotFound 错误
		return false, nil
	}
	return true, nil
}

// GetBucket 获取 Bucket 名称
func (c *S3Client) GetBucket() string {
	return c.config.Bucket
}

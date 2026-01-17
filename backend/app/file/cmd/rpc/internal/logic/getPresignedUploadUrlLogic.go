package logic

import (
	"context"
	"errors"
	"fmt"
	"path"
	"time"

	"polaris-io/backend/app/file/cmd/rpc/internal/svc"
	"polaris-io/backend/app/file/cmd/rpc/pb"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetPresignedUploadUrlLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPresignedUploadUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPresignedUploadUrlLogic {
	return &GetPresignedUploadUrlLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetPresignedUploadUrl 获取预签名上传 URL
// 生成 S3 Key 和预签名 URL，客户端使用此 URL 直接上传文件到 S3
func (l *GetPresignedUploadUrlLogic) GetPresignedUploadUrl(in *pb.GetPresignedUploadUrlReq) (*pb.GetPresignedUploadUrlResp, error) {
	// 参数校验
	if in.Hash == "" || in.Size == 0 {
		return nil, errors.New("hash and size are required")
	}

	// 生成 S3 Key
	// 格式: uploads/{year}/{month}/{day}/{uuid}.{ext}
	now := time.Now()
	uid := uuid.New().String()
	ext := in.Ext
	if ext == "" {
		ext = path.Ext(in.Name)
		if len(ext) > 0 && ext[0] == '.' {
			ext = ext[1:] // 去掉前导点
		}
	}

	var s3Key string
	if ext != "" {
		s3Key = fmt.Sprintf("uploads/%d/%02d/%02d/%s.%s",
			now.Year(), now.Month(), now.Day(), uid, ext)
	} else {
		s3Key = fmt.Sprintf("uploads/%d/%02d/%02d/%s",
			now.Year(), now.Month(), now.Day(), uid)
	}

	// 获取预签名上传 URL，有效期 1 小时
	uploadUrl, err := l.svcCtx.S3Client.GetPresignedUploadURL(l.ctx, s3Key, in.MimeType, 3600)
	if err != nil {
		l.Logger.Errorf("GetPresignedUploadUrl error: %v", err)
		return nil, fmt.Errorf("failed to generate presigned upload url: %w", err)
	}

	return &pb.GetPresignedUploadUrlResp{
		UploadUrl: uploadUrl,
		S3Key:     s3Key,
	}, nil
}

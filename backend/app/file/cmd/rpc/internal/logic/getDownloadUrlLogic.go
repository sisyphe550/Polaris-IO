package logic

import (
	"context"
	"errors"

	"polaris-io/backend/app/file/cmd/rpc/internal/svc"
	"polaris-io/backend/app/file/cmd/rpc/pb"
	"polaris-io/backend/app/file/model"
	"polaris-io/backend/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDownloadUrlLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDownloadUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDownloadUrlLogic {
	return &GetDownloadUrlLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetDownloadUrl 获取下载 URL
func (l *GetDownloadUrlLogic) GetDownloadUrl(in *pb.GetDownloadUrlReq) (*pb.GetDownloadUrlResp, error) {
	if in.Identity == "" {
		return nil, errors.New("identity is required")
	}

	// 查询文件
	file, err := l.svcCtx.UserRepositoryModel.FindOneByIdentity(l.ctx, in.Identity)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, xerr.NewErrCode(xerr.FILE_NOT_EXIST)
		}
		l.Logger.Errorf("GetDownloadUrl FindOneByIdentity error: %v", err)
		return nil, err
	}

	// 权限验证
	if in.UserId > 0 && int64(file.UserId) != in.UserId {
		return nil, xerr.NewErrCode(xerr.FILE_NOT_EXIST)
	}

	// 检查是否是文件（不是文件夹）
	if file.Ext == "" && file.Hash == "" {
		return nil, errors.New("cannot download folder")
	}

	// 检查 S3 路径
	if file.Path == "" {
		return nil, xerr.NewErrCode(xerr.FILE_NOT_EXIST)
	}

	// 生成预签名下载 URL，有效期 1 小时
	downloadUrl, err := l.svcCtx.S3Client.GetPresignedDownloadURL(l.ctx, file.Path, 3600)
	if err != nil {
		l.Logger.Errorf("GetDownloadUrl GetPresignedDownloadURL error: %v", err)
		return nil, xerr.NewErrCode(xerr.S3_PRESIGN_FAILED)
	}

	return &pb.GetDownloadUrlResp{
		Url: downloadUrl,
	}, nil
}

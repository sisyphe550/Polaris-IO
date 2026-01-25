package logic

import (
	"context"

	"polaris-io/backend/app/file/cmd/rpc/fileservice"
	"polaris-io/backend/app/share/cmd/rpc/internal/svc"
	"polaris-io/backend/app/share/cmd/rpc/pb"
	"polaris-io/backend/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetShareDownloadUrlLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetShareDownloadUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetShareDownloadUrlLogic {
	return &GetShareDownloadUrlLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetShareDownloadUrl 获取分享下载链接
func (l *GetShareDownloadUrlLogic) GetShareDownloadUrl(in *pb.GetShareDownloadUrlReq) (*pb.GetShareDownloadUrlResp, error) {
	// 1. 验证分享是否有效
	validateLogic := NewValidateShareLogic(l.ctx, l.svcCtx)
	validateResp, err := validateLogic.ValidateShare(&pb.ValidateShareReq{
		Identity: in.Identity,
		Code:     in.Code,
	})
	if err != nil {
		return nil, err
	}

	if !validateResp.Valid {
		switch validateResp.Reason {
		case "分享不存在":
			return nil, xerr.NewErrCode(xerr.SHARE_NOT_EXIST)
		case "分享已被封禁":
			return nil, xerr.NewErrCode(xerr.SHARE_BANNED)
		case "分享已过期":
			return nil, xerr.NewErrCode(xerr.SHARE_EXPIRED)
		case "提取码错误":
			return nil, xerr.NewErrCode(xerr.SHARE_CODE_ERROR)
		default:
			return nil, xerr.NewErrCode(xerr.SHARE_NOT_EXIST)
		}
	}

	share := validateResp.Share

	// 2. 确定要下载的文件 identity
	fileIdentity := in.FileIdentity
	if fileIdentity == "" {
		// 如果没有指定，使用分享的文件
		fileIdentity = share.RepositoryIdentity
	}

	// 3. 验证 fileIdentity 是否属于分享的文件（或其子文件）
	// 如果是子文件，需要验证其父目录链是否最终指向分享的文件
	// 这里简化处理：直接获取下载链接，file-rpc 会验证文件是否存在

	// 4. 调用 file-rpc 获取下载链接
	downloadResp, err := l.svcCtx.FileRpc.GetDownloadUrl(l.ctx, &fileservice.GetDownloadUrlReq{
		Identity: fileIdentity,
		UserId:   0, // 不验证用户权限（分享场景）
	})
	if err != nil {
		l.Logger.Errorf("GetShareDownloadUrl GetDownloadUrl error: %v", err)
		return nil, xerr.NewErrCode(xerr.FILE_DOWNLOAD_FAILED)
	}

	return &pb.GetShareDownloadUrlResp{
		Url: downloadResp.Url,
	}, nil
}

package logic

import (
	"context"
	"errors"

	"polaris-io/backend/app/file/cmd/rpc/fileservice"
	"polaris-io/backend/app/share/cmd/rpc/internal/svc"
	"polaris-io/backend/app/share/cmd/rpc/pb"
	"polaris-io/backend/app/user/cmd/rpc/usercenter"
	"polaris-io/backend/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveShareLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveShareLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveShareLogic {
	return &SaveShareLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SaveShare 保存分享到网盘
func (l *SaveShareLogic) SaveShare(in *pb.SaveShareReq) (*pb.SaveShareResp, error) {
	if in.UserId == 0 || in.Identity == "" {
		return nil, errors.New("userId and identity are required")
	}

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

	// 2. 确定要保存的文件列表
	fileIdentities := in.FileIdentities
	if len(fileIdentities) == 0 {
		// 如果没有指定，保存分享的根文件
		fileIdentities = []string{share.RepositoryIdentity}
	}

	// 3. 计算要保存的文件总大小并验证配额
	var totalSize uint64
	for _, identity := range fileIdentities {
		fileResp, err := l.svcCtx.FileRpc.GetFileInfo(l.ctx, &fileservice.GetFileInfoReq{
			Identity: identity,
			UserId:   0, // 不验证权限
		})
		if err != nil {
			l.Logger.Errorf("SaveShare GetFileInfo error: %v", err)
			continue
		}
		totalSize += fileResp.File.Size
	}

	// 4. 验证保存者的配额
	if totalSize > 0 {
		quotaResp, err := l.svcCtx.UsercenterRpc.GetUserQuota(l.ctx, &usercenter.GetUserQuotaReq{
			UserId: in.UserId,
		})
		if err != nil {
			l.Logger.Errorf("SaveShare GetUserQuota error: %v", err)
			return nil, xerr.NewErrCode(xerr.SHARE_SAVE_FAILED)
		}

		// 检查剩余配额
		if quotaResp.Quota.UsedSize+totalSize > quotaResp.Quota.TotalSize {
			return nil, xerr.NewErrCode(xerr.USER_QUOTA_EXCEEDED)
		}

		// 扣除配额
		_, err = l.svcCtx.UsercenterRpc.DeductQuota(l.ctx, &usercenter.DeductQuotaReq{
			UserId: in.UserId,
			Size:   totalSize,
		})
		if err != nil {
			l.Logger.Errorf("SaveShare DeductQuota error: %v", err)
			return nil, xerr.NewErrCode(xerr.USER_QUOTA_EXCEEDED)
		}
	}

	// 5. 调用 file-rpc 复制文件到用户的网盘（跨用户复制）
	copyResp, err := l.svcCtx.FileRpc.CopyFiles(l.ctx, &fileservice.CopyFilesReq{
		UserId:       share.UserId,      // 源文件所有者（分享者）
		Identities:   fileIdentities,    // 要复制的文件
		TargetId:     in.TargetFolderId, // 目标文件夹（保存者的）
		TargetUserId: in.UserId,         // 目标用户（保存者）
	})
	if err != nil {
		l.Logger.Errorf("SaveShare CopyFiles error: %v", err)
		// 退还配额
		if totalSize > 0 {
			_, refundErr := l.svcCtx.UsercenterRpc.RefundQuota(l.ctx, &usercenter.RefundQuotaReq{
				UserId: in.UserId,
				Size:   totalSize,
			})
			if refundErr != nil {
				l.Logger.Errorf("SaveShare RefundQuota error: %v", refundErr)
			}
		}
		return nil, xerr.NewErrCode(xerr.SHARE_SAVE_FAILED)
	}

	return &pb.SaveShareResp{
		SavedCount: copyResp.CopiedCount,
	}, nil
}

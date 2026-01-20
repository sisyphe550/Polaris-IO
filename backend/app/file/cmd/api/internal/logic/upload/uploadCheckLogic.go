package upload

import (
	"context"

	"polaris-io/backend/app/file/cmd/api/internal/svc"
	"polaris-io/backend/app/file/cmd/api/internal/types"
	"polaris-io/backend/app/file/cmd/rpc/fileservice"
	"polaris-io/backend/app/user/cmd/rpc/usercenter"
	"polaris-io/backend/pkg/ctxdata"
	"polaris-io/backend/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadCheckLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUploadCheckLogic 秒传检查 - 检查文件是否已存在
func NewUploadCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadCheckLogic {
	return &UploadCheckLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadCheckLogic) UploadCheck(req *types.UploadCheckReq) (resp *types.UploadCheckResp, err error) {
	userId := ctxdata.GetUidFromCtx(l.ctx)

	// 调用 RPC 检查秒传
	rpcResp, err := l.svcCtx.FileRpc.CheckInstantUpload(l.ctx, &fileservice.CheckInstantUploadReq{
		Hash: req.Hash,
		Size: req.Size,
	})
	if err != nil {
		return nil, err
	}

	// 如果秒传成功，需要创建用户文件记录
	if rpcResp.Exists && rpcResp.Meta != nil {
		// 1. 先扣减用户配额（秒传也需要占用用户空间）
		_, err = l.svcCtx.UsercenterRpc.DeductQuota(l.ctx, &usercenter.DeductQuotaReq{
			UserId: userId,
			Size:   req.Size,
		})
		if err != nil {
			l.Logger.Errorf("UploadCheck DeductQuota error: %v", err)
			return nil, xerr.NewErrCode(xerr.USER_QUOTA_EXCEEDED)
		}

		// 2. 创建文件记录（秒传）
		createResp, err := l.svcCtx.FileRpc.CreateFile(l.ctx, &fileservice.CreateFileReq{
			UserId:   userId,
			ParentId: 0, // 默认根目录，可以后续移动
			Name:     req.Name,
			Ext:      req.Ext,
			Hash:     req.Hash,
			Size:     req.Size,
			S3Key:    rpcResp.Meta.S3Key,
			MimeType: rpcResp.Meta.MimeType,
		})
		if err != nil {
			// 3. 创建文件记录失败，回滚配额
			_, _ = l.svcCtx.UsercenterRpc.RefundQuota(l.ctx, &usercenter.RefundQuotaReq{
				UserId: userId,
				Size:   req.Size,
			})
			return nil, err
		}

		return &types.UploadCheckResp{
			Exists:   true,
			Identity: createResp.Identity,
		}, nil
	}

	return &types.UploadCheckResp{
		Exists:   false,
		Identity: "",
	}, nil
}

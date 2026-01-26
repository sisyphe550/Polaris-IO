package upload

import (
	"context"
	"time"

	"polaris-io/backend/app/file/cmd/api/internal/svc"
	"polaris-io/backend/app/file/cmd/api/internal/types"
	"polaris-io/backend/app/file/cmd/rpc/fileservice"
	"polaris-io/backend/app/user/cmd/rpc/usercenter"
	"polaris-io/backend/pkg/asynqjob"
	"polaris-io/backend/pkg/ctxdata"
	"polaris-io/backend/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

// 上传超时时间 (24 小时)
const uploadTimeoutDuration = 24 * time.Hour

type UploadPresignLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUploadPresignLogic 获取预签名上传URL
func NewUploadPresignLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadPresignLogic {
	return &UploadPresignLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadPresignLogic) UploadPresign(req *types.UploadPresignReq) (resp *types.UploadPresignResp, err error) {
	userId := ctxdata.GetUidFromCtx(l.ctx)

	// 先检查用户配额是否足够
	_, err = l.svcCtx.UsercenterRpc.DeductQuota(l.ctx, &usercenter.DeductQuotaReq{
		UserId: userId,
		Size:   req.Size,
	})
	if err != nil {
		l.Logger.Errorf("UploadPresign DeductQuota error: %v", err)
		return nil, xerr.NewErrCode(xerr.USER_QUOTA_EXCEEDED)
	}

	// 获取预签名上传 URL
	rpcResp, err := l.svcCtx.FileRpc.GetPresignedUploadUrl(l.ctx, &fileservice.GetPresignedUploadUrlReq{
		UserId:   userId,
		Hash:     req.Hash,
		Size:     req.Size,
		Name:     req.Name,
		Ext:      req.Ext,
		MimeType: req.MimeType,
	})
	if err != nil {
		// 配额扣减失败回滚
		_, _ = l.svcCtx.UsercenterRpc.RefundQuota(l.ctx, &usercenter.RefundQuotaReq{
			UserId: userId,
			Size:   req.Size,
		})
		return nil, err
	}

	// 入队上传超时检测任务（延迟 24 小时执行）
	// 如果 24 小时后文件仍未上传完成，将自动退还配额
	if l.svcCtx.AsynqClient != nil {
		err = l.svcCtx.AsynqClient.EnqueueUploadTimeout(l.ctx, asynqjob.UploadTimeoutPayload{
			UserId:    userId,
			UploadKey: rpcResp.S3Key,
			Hash:      req.Hash,
			Size:      req.Size,
		}, uploadTimeoutDuration)
		if err != nil {
			// 入队失败只记录日志，不影响主流程
			l.Logger.Errorf("UploadPresign EnqueueUploadTimeout error: %v", err)
		}
	}

	// 使用 S3Key 作为 uploadKey，后续 complete 接口需要
	return &types.UploadPresignResp{
		UploadUrl: rpcResp.UploadUrl,
		UploadKey: rpcResp.S3Key,
	}, nil
}

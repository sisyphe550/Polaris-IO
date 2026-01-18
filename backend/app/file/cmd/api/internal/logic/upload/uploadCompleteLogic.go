package upload

import (
	"context"

	"polaris-io/backend/app/file/cmd/api/internal/svc"
	"polaris-io/backend/app/file/cmd/api/internal/types"
	"polaris-io/backend/app/file/cmd/rpc/fileservice"
	"polaris-io/backend/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadCompleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUploadCompleteLogic 上传完成回调
func NewUploadCompleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadCompleteLogic {
	return &UploadCompleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadCompleteLogic) UploadComplete(req *types.UploadCompleteReq) (resp *types.UploadCompleteResp, err error) {
	userId := ctxdata.GetUidFromCtx(l.ctx)

	// 创建文件记录
	rpcResp, err := l.svcCtx.FileRpc.CreateFile(l.ctx, &fileservice.CreateFileReq{
		UserId:   userId,
		ParentId: req.ParentId,
		Name:     req.Name,
		Ext:      req.Ext,
		Hash:     req.Hash,
		Size:     req.Size,
		S3Key:    req.UploadKey, // uploadKey 就是 S3Key
		MimeType: req.MimeType,
	})
	if err != nil {
		return nil, err
	}

	return &types.UploadCompleteResp{
		Identity: rpcResp.Identity,
	}, nil
}

package access

import (
	"context"

	"polaris-io/backend/app/share/cmd/api/internal/svc"
	"polaris-io/backend/app/share/cmd/api/internal/types"
	"polaris-io/backend/app/share/cmd/rpc/shareservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type ShareDownloadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewShareDownloadLogic 获取分享文件下载链接
func NewShareDownloadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShareDownloadLogic {
	return &ShareDownloadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ShareDownloadLogic) ShareDownload(req *types.ShareDownloadReq) (resp *types.ShareDownloadResp, err error) {
	// 调用 RPC 获取下载链接
	rpcResp, err := l.svcCtx.ShareRpc.GetShareDownloadUrl(l.ctx, &shareservice.GetShareDownloadUrlReq{
		Identity:     req.Identity,
		Code:         req.Code,
		FileIdentity: req.FileIdentity,
	})
	if err != nil {
		return nil, err
	}

	return &types.ShareDownloadResp{
		Url: rpcResp.Url,
	}, nil
}

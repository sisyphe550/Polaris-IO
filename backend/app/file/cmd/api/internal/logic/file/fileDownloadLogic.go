package file

import (
	"context"

	"polaris-io/backend/app/file/cmd/api/internal/svc"
	"polaris-io/backend/app/file/cmd/api/internal/types"
	"polaris-io/backend/app/file/cmd/rpc/fileservice"
	"polaris-io/backend/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type FileDownloadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFileDownloadLogic 获取下载URL
func NewFileDownloadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileDownloadLogic {
	return &FileDownloadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FileDownloadLogic) FileDownload(req *types.FileDownloadReq) (resp *types.FileDownloadResp, err error) {
	userId := ctxdata.GetUidFromCtx(l.ctx)

	rpcResp, err := l.svcCtx.FileRpc.GetDownloadUrl(l.ctx, &fileservice.GetDownloadUrlReq{
		Identity: req.Identity,
		UserId:   userId,
	})
	if err != nil {
		return nil, err
	}

	return &types.FileDownloadResp{
		Url: rpcResp.Url,
	}, nil
}

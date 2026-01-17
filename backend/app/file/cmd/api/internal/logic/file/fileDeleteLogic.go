package file

import (
	"context"

	"polaris-io/backend/app/file/cmd/api/internal/svc"
	"polaris-io/backend/app/file/cmd/api/internal/types"
	"polaris-io/backend/app/file/cmd/rpc/fileservice"
	"polaris-io/backend/app/user/cmd/rpc/usercenter"
	"polaris-io/backend/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type FileDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFileDeleteLogic 删除文件/文件夹 (移入回收站)
func NewFileDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileDeleteLogic {
	return &FileDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FileDeleteLogic) FileDelete(req *types.FileDeleteReq) (resp *types.FileDeleteResp, err error) {
	userId := ctxdata.GetUidFromCtx(l.ctx)

	rpcResp, err := l.svcCtx.FileRpc.SoftDeleteFiles(l.ctx, &fileservice.SoftDeleteFilesReq{
		UserId:     userId,
		Identities: req.Identities,
	})
	if err != nil {
		return nil, err
	}

	// 退还配额
	if rpcResp.FreedSize > 0 {
		_, err = l.svcCtx.UsercenterRpc.RefundQuota(l.ctx, &usercenter.RefundQuotaReq{
			UserId: userId,
			Size:   rpcResp.FreedSize,
		})
		if err != nil {
			l.Logger.Errorf("FileDelete RefundQuota error: %v", err)
			// 不影响主流程
		}
	}

	return &types.FileDeleteResp{}, nil
}

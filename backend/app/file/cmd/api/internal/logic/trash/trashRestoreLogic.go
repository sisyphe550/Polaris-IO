package trash

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

type TrashRestoreLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewTrashRestoreLogic 恢复文件
func NewTrashRestoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TrashRestoreLogic {
	return &TrashRestoreLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TrashRestoreLogic) TrashRestore(req *types.TrashRestoreReq) (resp *types.TrashRestoreResp, err error) {
	userId := ctxdata.GetUidFromCtx(l.ctx)

	rpcResp, err := l.svcCtx.FileRpc.RestoreFiles(l.ctx, &fileservice.RestoreFilesReq{
		UserId:     userId,
		Identities: req.Identities,
	})
	if err != nil {
		return nil, err
	}

	// 恢复后需要扣减配额
	if rpcResp.UsedSize > 0 {
		_, err = l.svcCtx.UsercenterRpc.DeductQuota(l.ctx, &usercenter.DeductQuotaReq{
			UserId: userId,
			Size:   rpcResp.UsedSize,
		})
		if err != nil {
			l.Logger.Errorf("TrashRestore DeductQuota error: %v", err)
			return nil, xerr.NewErrCode(xerr.USER_QUOTA_EXCEEDED)
		}
	}

	return &types.TrashRestoreResp{}, nil
}

package share

import (
	"context"

	"polaris-io/backend/app/share/cmd/api/internal/svc"
	"polaris-io/backend/app/share/cmd/api/internal/types"
	"polaris-io/backend/app/share/cmd/rpc/shareservice"
	"polaris-io/backend/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelShareLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCancelShareLogic 取消分享
func NewCancelShareLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelShareLogic {
	return &CancelShareLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelShareLogic) CancelShare(req *types.CancelShareReq) (resp *types.CancelShareResp, err error) {
	userId := ctxdata.GetUidFromCtx(l.ctx)

	rpcResp, err := l.svcCtx.ShareRpc.CancelShare(l.ctx, &shareservice.CancelShareReq{
		UserId:     userId,
		Identities: req.Identities,
	})
	if err != nil {
		return nil, err
	}

	return &types.CancelShareResp{
		CancelledCount: rpcResp.CancelledCount,
	}, nil
}

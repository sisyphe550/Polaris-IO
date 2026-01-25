package save

import (
	"context"

	"polaris-io/backend/app/share/cmd/api/internal/svc"
	"polaris-io/backend/app/share/cmd/api/internal/types"
	"polaris-io/backend/app/share/cmd/rpc/shareservice"
	"polaris-io/backend/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveShareLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewSaveShareLogic 保存分享到我的网盘
func NewSaveShareLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveShareLogic {
	return &SaveShareLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SaveShareLogic) SaveShare(req *types.SaveShareReq) (resp *types.SaveShareResp, err error) {
	userId := ctxdata.GetUidFromCtx(l.ctx)

	// 调用 RPC 保存分享
	rpcResp, err := l.svcCtx.ShareRpc.SaveShare(l.ctx, &shareservice.SaveShareReq{
		UserId:         userId,
		Identity:       req.Identity,
		Code:           req.Code,
		FileIdentities: req.FileIdentities,
		TargetFolderId: req.TargetFolderId,
	})
	if err != nil {
		return nil, err
	}

	return &types.SaveShareResp{
		SavedCount: rpcResp.SavedCount,
	}, nil
}

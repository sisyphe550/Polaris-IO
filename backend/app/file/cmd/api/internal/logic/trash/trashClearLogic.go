package trash

import (
	"context"

	"polaris-io/backend/app/file/cmd/api/internal/svc"
	"polaris-io/backend/app/file/cmd/api/internal/types"
	"polaris-io/backend/app/file/cmd/rpc/fileservice"
	"polaris-io/backend/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type TrashClearLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewTrashClearLogic 清空回收站
func NewTrashClearLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TrashClearLogic {
	return &TrashClearLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TrashClearLogic) TrashClear(req *types.TrashClearReq) (resp *types.TrashClearResp, err error) {
	userId := ctxdata.GetUidFromCtx(l.ctx)

	_, err = l.svcCtx.FileRpc.ClearTrash(l.ctx, &fileservice.ClearTrashReq{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}

	return &types.TrashClearResp{}, nil
}

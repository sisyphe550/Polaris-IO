package trash

import (
	"context"

	"polaris-io/backend/app/file/cmd/api/internal/svc"
	"polaris-io/backend/app/file/cmd/api/internal/types"
	"polaris-io/backend/app/file/cmd/rpc/fileservice"
	"polaris-io/backend/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type TrashDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewTrashDeleteLogic 彻底删除
func NewTrashDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TrashDeleteLogic {
	return &TrashDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TrashDeleteLogic) TrashDelete(req *types.TrashDeleteReq) (resp *types.TrashDeleteResp, err error) {
	userId := ctxdata.GetUidFromCtx(l.ctx)

	_, err = l.svcCtx.FileRpc.HardDeleteFiles(l.ctx, &fileservice.HardDeleteFilesReq{
		UserId:     userId,
		Identities: req.Identities,
	})
	if err != nil {
		return nil, err
	}

	return &types.TrashDeleteResp{}, nil
}

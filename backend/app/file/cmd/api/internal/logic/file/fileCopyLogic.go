package file

import (
	"context"

	"polaris-io/backend/app/file/cmd/api/internal/svc"
	"polaris-io/backend/app/file/cmd/api/internal/types"
	"polaris-io/backend/app/file/cmd/rpc/fileservice"
	"polaris-io/backend/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type FileCopyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFileCopyLogic 复制文件/文件夹
func NewFileCopyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileCopyLogic {
	return &FileCopyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FileCopyLogic) FileCopy(req *types.FileCopyReq) (resp *types.FileCopyResp, err error) {
	userId := ctxdata.GetUidFromCtx(l.ctx)

	_, err = l.svcCtx.FileRpc.CopyFiles(l.ctx, &fileservice.CopyFilesReq{
		UserId:     userId,
		Identities: req.Identities,
		TargetId:   req.TargetId,
	})
	if err != nil {
		return nil, err
	}

	return &types.FileCopyResp{}, nil
}

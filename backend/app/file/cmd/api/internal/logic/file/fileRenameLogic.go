package file

import (
	"context"

	"polaris-io/backend/app/file/cmd/api/internal/svc"
	"polaris-io/backend/app/file/cmd/api/internal/types"
	"polaris-io/backend/app/file/cmd/rpc/fileservice"
	"polaris-io/backend/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type FileRenameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFileRenameLogic 重命名文件/文件夹
func NewFileRenameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileRenameLogic {
	return &FileRenameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FileRenameLogic) FileRename(req *types.FileRenameReq) (resp *types.FileRenameResp, err error) {
	userId := ctxdata.GetUidFromCtx(l.ctx)

	_, err = l.svcCtx.FileRpc.RenameFile(l.ctx, &fileservice.RenameFileReq{
		UserId:   userId,
		Identity: req.Identity,
		Name:     req.Name,
	})
	if err != nil {
		return nil, err
	}

	return &types.FileRenameResp{}, nil
}

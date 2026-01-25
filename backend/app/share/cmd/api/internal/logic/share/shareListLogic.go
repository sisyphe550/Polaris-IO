package share

import (
	"context"

	"polaris-io/backend/app/file/cmd/rpc/fileservice"
	"polaris-io/backend/app/share/cmd/api/internal/svc"
	"polaris-io/backend/app/share/cmd/api/internal/types"
	"polaris-io/backend/app/share/cmd/rpc/shareservice"
	"polaris-io/backend/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type ShareListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewShareListLogic 我的分享列表
func NewShareListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShareListLogic {
	return &ShareListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ShareListLogic) ShareList(req *types.ShareListReq) (resp *types.ShareListResp, err error) {
	userId := ctxdata.GetUidFromCtx(l.ctx)

	// 调用 RPC 获取分享列表
	rpcResp, err := l.svcCtx.ShareRpc.GetShareList(l.ctx, &shareservice.GetShareListReq{
		UserId:   userId,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	// 转换响应，并获取文件信息
	list := make([]types.ShareItem, 0, len(rpcResp.List))
	for _, share := range rpcResp.List {
		item := types.ShareItem{
			Id:                 share.Id,
			Identity:           share.Identity,
			UserId:             share.UserId,
			RepositoryIdentity: share.RepositoryIdentity,
			Code:               share.Code,
			ClickNum:           share.ClickNum,
			ExpiredTime:        share.ExpiredTime,
			Status:             share.Status,
			CreateTime:         share.CreateTime,
		}

		// 获取文件信息
		fileResp, err := l.svcCtx.FileRpc.GetFileInfo(l.ctx, &fileservice.GetFileInfoReq{
			Identity: share.RepositoryIdentity,
			UserId:   userId,
		})
		if err == nil && fileResp.File != nil {
			item.FileName = fileResp.File.Name
			item.FileExt = fileResp.File.Ext
			item.FileSize = fileResp.File.Size
			item.IsDir = fileResp.File.IsDir
		}

		list = append(list, item)
	}

	return &types.ShareListResp{
		List:  list,
		Total: rpcResp.Total,
	}, nil
}

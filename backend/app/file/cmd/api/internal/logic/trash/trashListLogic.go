package trash

import (
	"context"

	"polaris-io/backend/app/file/cmd/api/internal/svc"
	"polaris-io/backend/app/file/cmd/api/internal/types"
	"polaris-io/backend/app/file/cmd/rpc/fileservice"
	"polaris-io/backend/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type TrashListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewTrashListLogic 回收站列表
func NewTrashListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TrashListLogic {
	return &TrashListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TrashListLogic) TrashList(req *types.TrashListReq) (resp *types.TrashListResp, err error) {
	userId := ctxdata.GetUidFromCtx(l.ctx)

	rpcResp, err := l.svcCtx.FileRpc.ListTrash(l.ctx, &fileservice.ListTrashReq{
		UserId:   userId,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	// 转换响应
	list := make([]types.TrashItem, 0, len(rpcResp.List))
	for _, f := range rpcResp.List {
		list = append(list, types.TrashItem{
			Id:         f.Id,
			Identity:   f.Identity,
			Name:       f.Name,
			Ext:        f.Ext,
			Size:       f.Size,
			IsDir:      f.IsDir,
			DeleteTime: f.UpdateTime, // 使用 UpdateTime 作为删除时间
		})
	}

	return &types.TrashListResp{
		List:  list,
		Total: rpcResp.Total,
	}, nil
}

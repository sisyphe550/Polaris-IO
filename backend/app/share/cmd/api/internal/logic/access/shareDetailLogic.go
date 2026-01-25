package access

import (
	"context"

	"polaris-io/backend/app/share/cmd/api/internal/svc"
	"polaris-io/backend/app/share/cmd/api/internal/types"
	"polaris-io/backend/app/share/cmd/rpc/shareservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type ShareDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewShareDetailLogic 获取分享详情（验证提取码）
func NewShareDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShareDetailLogic {
	return &ShareDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ShareDetailLogic) ShareDetail(req *types.ShareDetailReq) (resp *types.ShareDetailResp, err error) {
	// 调用 RPC 获取分享详情
	rpcResp, err := l.svcCtx.ShareRpc.GetShareDetail(l.ctx, &shareservice.GetShareDetailReq{
		Identity: req.Identity,
		Code:     req.Code,
	})
	if err != nil {
		return nil, err
	}

	// 转换文件信息
	file := types.ShareFileInfo{
		Identity: rpcResp.File.Identity,
		Name:     rpcResp.File.Name,
		Ext:      rpcResp.File.Ext,
		Size:     rpcResp.File.Size,
		IsDir:    rpcResp.File.IsDir,
	}

	// 转换子文件列表
	children := make([]types.ShareFileInfo, 0, len(rpcResp.Children))
	for _, child := range rpcResp.Children {
		children = append(children, types.ShareFileInfo{
			Identity: child.Identity,
			Name:     child.Name,
			Ext:      child.Ext,
			Size:     child.Size,
			IsDir:    child.IsDir,
		})
	}

	return &types.ShareDetailResp{
		File:        file,
		SharerName:  rpcResp.SharerName,
		ExpiredTime: rpcResp.Share.ExpiredTime,
		Children:    children,
	}, nil
}

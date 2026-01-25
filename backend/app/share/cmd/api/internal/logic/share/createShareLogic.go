package share

import (
	"context"
	"fmt"

	"polaris-io/backend/app/share/cmd/api/internal/svc"
	"polaris-io/backend/app/share/cmd/api/internal/types"
	"polaris-io/backend/app/share/cmd/rpc/shareservice"
	"polaris-io/backend/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateShareLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreateShareLogic 创建分享
func NewCreateShareLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateShareLogic {
	return &CreateShareLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateShareLogic) CreateShare(req *types.CreateShareReq) (resp *types.CreateShareResp, err error) {
	userId := ctxdata.GetUidFromCtx(l.ctx)

	// 调用 RPC
	rpcResp, err := l.svcCtx.ShareRpc.CreateShare(l.ctx, &shareservice.CreateShareReq{
		UserId:             userId,
		RepositoryIdentity: req.RepositoryIdentity,
		ExpiredType:        req.ExpiredType,
		HasCode:            req.HasCode,
	})
	if err != nil {
		return nil, err
	}

	// 生成分享链接
	shareUrl := fmt.Sprintf("%s?identity=%s", l.svcCtx.Config.Share.BaseUrl, rpcResp.Identity)
	if rpcResp.Code != "" {
		shareUrl += "&code=" + rpcResp.Code
	}

	return &types.CreateShareResp{
		Identity: rpcResp.Identity,
		Code:     rpcResp.Code,
		Url:      shareUrl,
	}, nil
}

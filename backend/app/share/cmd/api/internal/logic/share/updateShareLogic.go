package share

import (
	"context"

	"polaris-io/backend/app/share/cmd/api/internal/svc"
	"polaris-io/backend/app/share/cmd/api/internal/types"
	"polaris-io/backend/app/share/cmd/rpc/shareservice"
	"polaris-io/backend/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateShareLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateShareLogic 更新分享
func NewUpdateShareLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateShareLogic {
	return &UpdateShareLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateShareLogic) UpdateShare(req *types.UpdateShareReq) (resp *types.UpdateShareResp, err error) {
	userId := ctxdata.GetUidFromCtx(l.ctx)

	_, err = l.svcCtx.ShareRpc.UpdateShare(l.ctx, &shareservice.UpdateShareReq{
		UserId:      userId,
		Identity:    req.Identity,
		ExpiredType: req.ExpiredType,
		Code:        req.Code,
		UpdateCode:  req.Code != "", // 如果提供了 code 参数则更新
	})
	if err != nil {
		return nil, err
	}

	return &types.UpdateShareResp{}, nil
}

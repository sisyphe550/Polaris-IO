package logic

import (
	"context"

	"polaris-io/backend/app/share/cmd/rpc/internal/svc"
	"polaris-io/backend/app/share/cmd/rpc/pb"
	"polaris-io/backend/app/share/model"
	"polaris-io/backend/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetShareByIdentityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetShareByIdentityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetShareByIdentityLogic {
	return &GetShareByIdentityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetShareByIdentity 根据标识获取分享信息
func (l *GetShareByIdentityLogic) GetShareByIdentity(in *pb.GetShareByIdentityReq) (*pb.GetShareByIdentityResp, error) {
	share, err := l.svcCtx.ShareModel.FindOneByIdentity(l.ctx, in.Identity)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, xerr.NewErrCode(xerr.SHARE_NOT_EXIST)
		}
		l.Logger.Errorf("GetShareByIdentity FindOneByIdentity error: %v", err)
		return nil, err
	}

	return &pb.GetShareByIdentityResp{
		Share: convertShareToProto(share),
	}, nil
}

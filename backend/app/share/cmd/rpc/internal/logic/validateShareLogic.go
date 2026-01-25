package logic

import (
	"context"
	"time"

	"polaris-io/backend/app/share/cmd/rpc/internal/svc"
	"polaris-io/backend/app/share/cmd/rpc/pb"
	"polaris-io/backend/app/share/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidateShareLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateShareLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateShareLogic {
	return &ValidateShareLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ValidateShare 验证分享是否有效
func (l *ValidateShareLogic) ValidateShare(in *pb.ValidateShareReq) (*pb.ValidateShareResp, error) {
	// 1. 根据 identity 查询分享记录
	share, err := l.svcCtx.ShareModel.FindOneByIdentity(l.ctx, in.Identity)
	if err != nil {
		if err == model.ErrNotFound {
			return &pb.ValidateShareResp{
				Valid:  false,
				Reason: "分享不存在",
			}, nil
		}
		l.Logger.Errorf("ValidateShare FindOneByIdentity error: %v", err)
		return &pb.ValidateShareResp{
			Valid:  false,
			Reason: "查询分享失败",
		}, nil
	}

	// 2. 检查是否被封禁
	if share.Status == 1 {
		return &pb.ValidateShareResp{
			Valid:  false,
			Reason: "分享已被封禁",
			Share:  convertShareToProto(share),
		}, nil
	}

	// 3. 检查是否过期
	if share.ExpiredTime > 0 && uint64(time.Now().Unix()) > share.ExpiredTime {
		return &pb.ValidateShareResp{
			Valid:  false,
			Reason: "分享已过期",
			Share:  convertShareToProto(share),
		}, nil
	}

	// 4. 检查提取码（如果设置了提取码）
	if share.Code != "" && share.Code != in.Code {
		return &pb.ValidateShareResp{
			Valid:  false,
			Reason: "提取码错误",
			Share:  convertShareToProto(share),
		}, nil
	}

	return &pb.ValidateShareResp{
		Valid: true,
		Share: convertShareToProto(share),
	}, nil
}

// convertShareToProto 将 model.Share 转换为 pb.ShareInfo
func convertShareToProto(share *model.Share) *pb.ShareInfo {
	if share == nil {
		return nil
	}
	return &pb.ShareInfo{
		Id:                 int64(share.Id),
		Identity:           share.Identity,
		UserId:             int64(share.UserId),
		RepositoryIdentity: share.RepositoryIdentity,
		Code:               share.Code,
		ClickNum:           int64(share.ClickNum),
		ExpiredTime:        int64(share.ExpiredTime),
		Status:             share.Status,
		CreateTime:         share.CreateTime.Unix(),
		UpdateTime:         share.UpdateTime.Unix(),
	}
}

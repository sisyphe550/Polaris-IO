package logic

import (
	"context"

	"polaris-io/backend/app/user/cmd/rpc/internal/svc"
	"polaris-io/backend/app/user/cmd/rpc/pb"
	"polaris-io/backend/app/user/model"
	"polaris-io/backend/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetUserInfo 获取用户信息
func (l *GetUserInfoLogic) GetUserInfo(in *pb.GetUserInfoReq) (*pb.GetUserInfoResp, error) {
	// 根据用户 ID 查询用户信息
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(in.UserId))
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.USER_NOT_EXIST), "userId:%d not found", in.UserId)
		}
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "FindOne err:%v, userId:%d", err, in.UserId)
	}

	return &pb.GetUserInfoResp{
		User: &pb.User{
			Id:     int64(user.Id),
			Mobile: user.Mobile,
			Name:   user.Name,
			Avatar: user.Avatar,
			Info:   user.Info,
		},
	}, nil
}

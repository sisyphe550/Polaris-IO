package logic

import (
	"context"

	"shared-board/backend/app/user/cmd/rpc/internal/svc"
	"shared-board/backend/app/user/cmd/rpc/pb"
	"shared-board/backend/app/user/model"
	"shared-board/backend/pkg/xerr"

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

// 获取信息
func (l *GetUserInfoLogic) GetUserInfo(in *pb.GetUserInfoReq) (*pb.GetUserInfoResp, error) {
	if in == nil || in.UserId <= 0 {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.REUQEST_PARAM_ERROR), "invalid userId")
	}

	uid := uint64(in.UserId)

	var u *model.Users
	if l.svcCtx.UsersModel != nil {
		user, err := l.svcCtx.UsersModel.FindOne(l.ctx, uid)
		if err != nil && err != model.ErrNotFound {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "FindOne err userId:%d, err:%v", in.UserId, err)
		}
		if user == nil {
			return nil, errors.Wrapf(ErrUserNoExistsError, "userId:%d", in.UserId)
		}
		u = user
	} else {
		user, ok := l.svcCtx.MemUsers.FindOne(uid)
		if !ok || user == nil {
			return nil, errors.Wrapf(ErrUserNoExistsError, "userId:%d", in.UserId)
		}
		u = user
	}

	// 当前 users 表字段与 proto(User) 不完全一致：
	// - 将 Users.Username 映射到 mobile
	// - nickname 暂时用 username 兜底（后续你如果补表字段再调整）
	return &pb.GetUserInfoResp{
		User: &pb.User{
			Id:       int64(u.Id),
			Mobile:   u.Username,
			Nickname: u.Username,
			Avatar:   u.Avatar,
			Info:     u.Info,
		},
	}, nil

}

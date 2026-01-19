package logic

import (
	"context"

	"polaris-io/backend/app/user/cmd/rpc/internal/svc"
	"polaris-io/backend/app/user/cmd/rpc/pb"
	"polaris-io/backend/app/user/model"
	"polaris-io/backend/pkg/tool"
	"polaris-io/backend/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 默认配额：10GB (字节)
const DefaultQuotaSize uint64 = 10 * 1024 * 1024 * 1024

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Register 用户注册 (同时创建 10GB 配额)
func (l *RegisterLogic) Register(in *pb.RegisterReq) (*pb.RegisterResp, error) {
	// 1. 检查手机号是否已注册
	_, err := l.svcCtx.UserModel.FindOneByMobile(l.ctx, in.Mobile)
	if err == nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.USER_ALREADY_EXISTS), "mobile:%s already registered", in.Mobile)
	}
	if err != model.ErrNotFound {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "FindOneByMobile err:%v, mobile:%s", err, in.Mobile)
	}

	// 2. 事务：创建用户和配额
	var userId int64
	err = l.svcCtx.UserModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 2.1 创建用户
		user := &model.User{
			Mobile:   in.Mobile,
			Password: tool.Md5ByString(in.Password),
			Name:     in.Name,
		}
		// 如果昵称为空，随机生成
		if len(user.Name) == 0 {
			user.Name = "用户" + tool.Krand(6, tool.KC_RAND_KIND_NUM)
		}

		insertResult, err := l.svcCtx.UserModel.Insert(ctx, session, user)
		if err != nil {
			return errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "Insert user err:%v, user:%+v", err, user)
		}
		lastId, err := insertResult.LastInsertId()
		if err != nil {
			return errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "LastInsertId err:%v", err)
		}
		userId = lastId

		// 2.2 创建配额记录 (默认 10GB)
		quota := &model.UserQuota{
			UserId:    uint64(userId),
			TotalSize: DefaultQuotaSize,
			UsedSize:  0,
		}
		_, err = l.svcCtx.UserQuotaModel.Insert(ctx, session, quota)
		if err != nil {
			return errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "Insert quota err:%v, quota:%+v", err, quota)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// 3. 预热配额缓存（可选，失败不影响主流程）
	_ = l.svcCtx.UserQuotaModel.WarmUpCache(l.ctx, uint64(userId))

	// 4. 生成 Token
	generateTokenLogic := NewGenerateTokenLogic(l.ctx, l.svcCtx)
	tokenResp, err := generateTokenLogic.GenerateToken(&pb.GenerateTokenReq{
		UserId: userId,
	})
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.TOKEN_GENERATE_ERROR), "GenerateToken err:%v, userId:%d", err, userId)
	}

	return &pb.RegisterResp{
		AccessToken:  tokenResp.AccessToken,
		AccessExpire: tokenResp.AccessExpire,
		RefreshAfter: tokenResp.RefreshAfter,
	}, nil
}

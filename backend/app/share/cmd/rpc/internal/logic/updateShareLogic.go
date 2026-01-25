package logic

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"polaris-io/backend/app/share/cmd/rpc/internal/svc"
	"polaris-io/backend/app/share/cmd/rpc/pb"
	"polaris-io/backend/app/share/model"
	"polaris-io/backend/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateShareLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateShareLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateShareLogic {
	return &UpdateShareLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateShare 更新分享
func (l *UpdateShareLogic) UpdateShare(in *pb.UpdateShareReq) (*pb.UpdateShareResp, error) {
	if in.UserId == 0 || in.Identity == "" {
		return nil, errors.New("userId and identity are required")
	}

	// 1. 查询分享记录
	share, err := l.svcCtx.ShareModel.FindOneByIdentity(l.ctx, in.Identity)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, xerr.NewErrCode(xerr.SHARE_NOT_EXIST)
		}
		l.Logger.Errorf("UpdateShare FindOneByIdentity error: %v", err)
		return nil, err
	}

	// 2. 验证权限
	if share.UserId != uint64(in.UserId) {
		return nil, xerr.NewErrCode(xerr.SHARE_PERMISSION_DENIED)
	}

	// 3. 更新字段
	needUpdate := false

	// 更新有效期
	if in.ExpiredType >= 0 {
		if in.ExpiredType == 0 {
			share.ExpiredTime = 0 // 永久有效
		} else {
			share.ExpiredTime = uint64(time.Now().Add(time.Duration(in.ExpiredType) * 24 * time.Hour).Unix())
		}
		needUpdate = true
	}

	// 更新提取码
	if in.UpdateCode {
		if in.Code == "" {
			share.Code = "" // 取消提取码
		} else if in.Code == "-1" {
			// 不修改
		} else {
			share.Code = in.Code // 设置新提取码
		}
		needUpdate = true
	}

	if !needUpdate {
		return &pb.UpdateShareResp{}, nil
	}

	// 4. 保存更新（使用乐观锁）
	if err := l.svcCtx.ShareModel.UpdateWithVersion(l.ctx, nil, share); err != nil {
		l.Logger.Errorf("UpdateShare UpdateWithVersion error: %v", err)
		return nil, err
	}

	return &pb.UpdateShareResp{}, nil
}

// generateRandomCode 生成随机提取码
func generateRandomCode(length int) string {
	if length <= 0 {
		length = 4
	}
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := make([]byte, length)
	for i := range code {
		code[i] = charset[r.Intn(len(charset))]
	}
	return string(code)
}

package logic

import (
	"context"
	"errors"

	"polaris-io/backend/app/file/cmd/rpc/internal/svc"
	"polaris-io/backend/app/file/cmd/rpc/pb"
	"polaris-io/backend/app/file/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckFileExistsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckFileExistsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckFileExistsLogic {
	return &CheckFileExistsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CheckFileExists 检查文件是否存在
func (l *CheckFileExistsLogic) CheckFileExists(in *pb.CheckFileExistsReq) (*pb.CheckFileExistsResp, error) {
	if in.Identity == "" {
		return &pb.CheckFileExistsResp{Exists: false}, nil
	}

	file, err := l.svcCtx.UserRepositoryModel.FindOneByIdentity(l.ctx, in.Identity)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &pb.CheckFileExistsResp{Exists: false}, nil
		}
		l.Logger.Errorf("CheckFileExists FindOneByIdentity error: %v", err)
		return nil, err
	}

	// 权限验证
	if in.UserId > 0 && int64(file.UserId) != in.UserId {
		return &pb.CheckFileExistsResp{Exists: false}, nil
	}

	return &pb.CheckFileExistsResp{Exists: true}, nil
}

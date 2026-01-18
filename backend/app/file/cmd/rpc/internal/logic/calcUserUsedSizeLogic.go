package logic

import (
	"context"

	"polaris-io/backend/app/file/cmd/rpc/internal/svc"
	"polaris-io/backend/app/file/cmd/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CalcUserUsedSizeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCalcUserUsedSizeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CalcUserUsedSizeLogic {
	return &CalcUserUsedSizeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CalcUserUsedSize 计算用户已用空间
func (l *CalcUserUsedSizeLogic) CalcUserUsedSize(in *pb.CalcUserUsedSizeReq) (*pb.CalcUserUsedSizeResp, error) {
	if in.UserId == 0 {
		return &pb.CalcUserUsedSizeResp{UsedSize: 0}, nil
	}

	// 计算用户所有未删除文件的总大小
	builder := l.svcCtx.UserRepositoryModel.SelectBuilder().
		Where("user_id = ?", in.UserId)

	usedSize, err := l.svcCtx.UserRepositoryModel.FindSum(l.ctx, builder, "size")
	if err != nil {
		l.Logger.Errorf("CalcUserUsedSize FindSum error: %v", err)
		return nil, err
	}

	return &pb.CalcUserUsedSizeResp{
		UsedSize: uint64(usedSize),
	}, nil
}

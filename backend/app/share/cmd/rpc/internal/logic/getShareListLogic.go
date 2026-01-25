package logic

import (
	"context"
	"errors"

	"polaris-io/backend/app/share/cmd/rpc/internal/svc"
	"polaris-io/backend/app/share/cmd/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetShareListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetShareListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetShareListLogic {
	return &GetShareListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetShareList 获取分享列表
func (l *GetShareListLogic) GetShareList(in *pb.GetShareListReq) (*pb.GetShareListResp, error) {
	if in.UserId == 0 {
		return nil, errors.New("userId is required")
	}

	// 分页参数处理
	page := in.Page
	pageSize := in.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 查询分享列表
	shares, total, err := l.svcCtx.ShareModel.FindListByUserId(l.ctx, uint64(in.UserId), page, pageSize)
	if err != nil {
		l.Logger.Errorf("GetShareList FindListByUserId error: %v", err)
		return nil, err
	}

	// 转换为响应格式
	list := make([]*pb.ShareInfo, 0, len(shares))
	for _, share := range shares {
		list = append(list, convertShareToProto(share))
	}

	return &pb.GetShareListResp{
		List:  list,
		Total: total,
	}, nil
}

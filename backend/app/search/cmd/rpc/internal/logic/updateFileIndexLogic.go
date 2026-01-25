package logic

import (
	"context"

	"polaris-io/backend/app/search/cmd/rpc/internal/svc"
	"polaris-io/backend/app/search/cmd/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateFileIndexLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateFileIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFileIndexLogic {
	return &UpdateFileIndexLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateFileIndex 更新文件索引
func (l *UpdateFileIndexLogic) UpdateFileIndex(in *pb.UpdateFileIndexReq) (*pb.UpdateFileIndexResp, error) {
	fields := make(map[string]interface{})

	if in.Name != "" {
		fields["name"] = in.Name
	}

	// 使用 updateParentId 标志来判断是否需要更新 parentId
	// 这样可以区分 "不更新" 和 "更新为 0（根目录）"
	if in.UpdateParentId {
		fields["parent_id"] = in.ParentId
	}

	if len(fields) == 0 {
		return &pb.UpdateFileIndexResp{}, nil
	}

	if err := l.svcCtx.ESClient.UpdateDocument(l.ctx, in.Id, fields); err != nil {
		l.Logger.Errorf("UpdateFileIndex error: %v", err)
		return nil, err
	}

	l.Logger.Infof("Updated file index: id=%s", in.Id)
	return &pb.UpdateFileIndexResp{}, nil
}

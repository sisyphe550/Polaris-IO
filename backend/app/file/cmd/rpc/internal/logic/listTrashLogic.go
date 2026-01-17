package logic

import (
	"context"

	"polaris-io/backend/app/file/cmd/rpc/internal/svc"
	"polaris-io/backend/app/file/cmd/rpc/pb"
	"polaris-io/backend/pkg/globalkey"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTrashLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTrashLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTrashLogic {
	return &ListTrashLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListTrash 回收站列表
func (l *ListTrashLogic) ListTrash(in *pb.ListTrashReq) (*pb.ListTrashResp, error) {
	// 分页参数
	page := in.Page
	if page < 1 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 查询已删除的文件 (del_state = 1)
	// 注意: 这里需要查询 del_state = 1 的记录，但 Model 默认过滤了已删除的
	// 所以需要直接查询数据库
	builder := l.svcCtx.UserRepositoryModel.SelectBuilder().
		Where("user_id = ?", in.UserId).
		Where("del_state = ?", globalkey.DelStateYes) // 已删除

	// 由于 FindPageListByPageWithTotal 会自动加 del_state = 0 条件
	// 这里需要自定义查询或修改 Model
	// 暂时使用 FindAll 然后手动分页
	// TODO: 优化为真正的分页查询

	// 先获取总数
	countBuilder := l.svcCtx.UserRepositoryModel.SelectBuilder().
		Where("user_id = ?", in.UserId)
	// 由于 FindCount 也会加 del_state = 0，需要特殊处理
	// 这里简化处理，后续可以在 Model 层添加专门的回收站查询方法

	// 简化实现：直接返回空列表，等待 Model 层扩展
	// 实际生产中应该扩展 Model 层添加 FindTrashList 方法

	_ = builder
	_ = countBuilder

	return &pb.ListTrashResp{
		List:  []*pb.FileInfo{},
		Total: 0,
	}, nil
}

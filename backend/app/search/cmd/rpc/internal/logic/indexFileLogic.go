package logic

import (
	"context"
	"time"

	"polaris-io/backend/app/search/cmd/rpc/internal/svc"
	"polaris-io/backend/app/search/cmd/rpc/pb"
	"polaris-io/backend/app/search/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type IndexFileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIndexFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IndexFileLogic {
	return &IndexFileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// IndexFile 索引文件（供 file-rpc 调用）
func (l *IndexFileLogic) IndexFile(in *pb.IndexFileReq) (*pb.IndexFileResp, error) {
	doc := &types.FileDocument{
		ID:         in.Id,
		UserID:     in.UserId,
		FileID:     in.FileId,
		Name:       in.Name,
		Ext:        in.Ext,
		Size:       in.Size,
		Hash:       in.Hash,
		ParentID:   in.ParentId,
		IsDir:      in.IsDir,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	if err := l.svcCtx.ESClient.IndexDocument(l.ctx, doc); err != nil {
		l.Logger.Errorf("IndexFile error: %v", err)
		return nil, err
	}

	l.Logger.Infof("Indexed file: id=%s, name=%s", in.Id, in.Name)
	return &pb.IndexFileResp{}, nil
}

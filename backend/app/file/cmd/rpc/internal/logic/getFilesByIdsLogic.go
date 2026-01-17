package logic

import (
	"context"
	"errors"

	"polaris-io/backend/app/file/cmd/rpc/internal/svc"
	"polaris-io/backend/app/file/cmd/rpc/pb"
	"polaris-io/backend/app/file/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFilesByIdsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFilesByIdsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFilesByIdsLogic {
	return &GetFilesByIdsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetFilesByIds 批量获取文件信息
func (l *GetFilesByIdsLogic) GetFilesByIds(in *pb.GetFilesByIdsReq) (*pb.GetFilesByIdsResp, error) {
	if len(in.Identities) == 0 {
		return &pb.GetFilesByIdsResp{Files: []*pb.FileInfo{}}, nil
	}

	files := make([]*pb.FileInfo, 0, len(in.Identities))

	for _, identity := range in.Identities {
		file, err := l.svcCtx.UserRepositoryModel.FindOneByIdentity(l.ctx, identity)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				continue
			}
			l.Logger.Errorf("GetFilesByIds FindOneByIdentity error: %v", err)
			continue
		}

		// 权限验证
		if in.UserId > 0 && int64(file.UserId) != in.UserId {
			continue
		}

		isDir := file.Ext == "" && file.Hash == ""
		files = append(files, &pb.FileInfo{
			Id:         int64(file.Id),
			Identity:   file.Identity,
			Hash:       file.Hash,
			UserId:     int64(file.UserId),
			ParentId:   int64(file.ParentId),
			Name:       file.Name,
			Ext:        file.Ext,
			Size:       file.Size,
			Path:       file.Path,
			IsDir:      isDir,
			CreateTime: file.CreateTime.Unix(),
			UpdateTime: file.UpdateTime.Unix(),
		})
	}

	return &pb.GetFilesByIdsResp{Files: files}, nil
}

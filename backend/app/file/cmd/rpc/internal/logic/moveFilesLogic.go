package logic

import (
	"context"
	"errors"

	"polaris-io/backend/app/file/cmd/rpc/internal/svc"
	"polaris-io/backend/app/file/cmd/rpc/pb"
	"polaris-io/backend/app/file/model"
	"polaris-io/backend/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type MoveFilesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMoveFilesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MoveFilesLogic {
	return &MoveFilesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// MoveFiles 移动文件/文件夹
func (l *MoveFilesLogic) MoveFiles(in *pb.MoveFilesReq) (*pb.MoveFilesResp, error) {
	if in.UserId == 0 || len(in.Identities) == 0 {
		return nil, errors.New("userId and identities are required")
	}

	// 检查目标目录是否存在（如果不是根目录）
	if in.TargetId > 0 {
		targetFile, err := l.svcCtx.UserRepositoryModel.FindOne(l.ctx, uint64(in.TargetId))
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil, xerr.NewErrCode(xerr.FILE_PARENT_NOT_EXIST)
			}
			return nil, err
		}
		// 检查目标是文件夹
		if targetFile.Ext != "" || targetFile.Hash != "" {
			return nil, xerr.NewErrCode(xerr.FILE_PARENT_NOT_EXIST)
		}
		// 检查权限
		if int64(targetFile.UserId) != in.UserId {
			return nil, xerr.NewErrCode(xerr.FILE_PARENT_NOT_EXIST)
		}
	}

	var movedCount int64

	for _, identity := range in.Identities {
		// 查询文件
		file, err := l.svcCtx.UserRepositoryModel.FindOneByIdentity(l.ctx, identity)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				continue
			}
			l.Logger.Errorf("MoveFiles FindOneByIdentity error: %v", err)
			continue
		}

		// 权限验证
		if int64(file.UserId) != in.UserId {
			continue
		}

		// 检查不能移动到自身或子目录（如果是文件夹）
		if file.Ext == "" && file.Hash == "" {
			if int64(file.Id) == in.TargetId {
				return nil, xerr.NewErrCode(xerr.FILE_CANNOT_MOVE_TO_SELF)
			}
			// TODO: 检查是否移动到子目录
		}

		// 更新 parent_id
		file.ParentId = uint64(in.TargetId)
		_, err = l.svcCtx.UserRepositoryModel.Update(l.ctx, nil, file)
		if err != nil {
			l.Logger.Errorf("MoveFiles Update error: %v", err)
			continue
		}

		movedCount++

		// 发送 Kafka 更新事件
		if err := l.svcCtx.KafkaProducer.SendFileUpdated(
			l.ctx, in.UserId, int64(file.Id), identity, file.Name, in.TargetId); err != nil {
			l.Logger.Errorf("MoveFiles SendFileUpdated error: %v", err)
		}
	}

	return &pb.MoveFilesResp{
		MovedCount: movedCount,
	}, nil
}

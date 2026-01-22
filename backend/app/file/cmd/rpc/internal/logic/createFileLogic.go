package logic

import (
	"context"
	"errors"

	"polaris-io/backend/app/file/cmd/rpc/internal/svc"
	"polaris-io/backend/app/file/cmd/rpc/pb"
	"polaris-io/backend/app/file/model"
	fileMongo "polaris-io/backend/app/file/mongo"
	"polaris-io/backend/pkg/xerr"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateFileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateFileLogic {
	return &CreateFileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateFile 创建文件记录 (上传完成后调用)
// 1. 创建/更新 MongoDB file_meta (如果是秒传则增加引用计数)
// 2. 创建 MySQL user_repository 记录
// 3. 发送 Kafka 事件
func (l *CreateFileLogic) CreateFile(in *pb.CreateFileReq) (*pb.CreateFileResp, error) {
	// 参数校验
	if in.UserId == 0 || in.Hash == "" || in.Name == "" {
		return nil, errors.New("userId, hash and name are required")
	}

	// 生成文件唯一标识
	identity := uuid.New().String()

	// 1. 处理 MongoDB file_meta
	// 检查是否已存在（秒传场景）
	existingMeta, err := l.svcCtx.FileMetaModel.FindByHash(l.ctx, in.Hash)
	if err != nil && !errors.Is(err, fileMongo.ErrNotFound) {
		l.Logger.Errorf("CreateFile FindByHash error: %v", err)
		return nil, err
	}

	if existingMeta != nil {
		// 已存在，增加引用计数
		if err := l.svcCtx.FileMetaModel.IncrRefCount(l.ctx, in.Hash, 1); err != nil {
			l.Logger.Errorf("CreateFile IncrRefCount error: %v", err)
			return nil, err
		}
	} else {
		// 不存在，创建新记录
		newMeta := &fileMongo.FileMeta{
			Hash:     in.Hash,
			Size:     in.Size,
			S3Key:    in.S3Key,
			Ext:      in.Ext,
			MimeType: in.MimeType,
			RefCount: 1,
		}
		if err := l.svcCtx.FileMetaModel.Insert(l.ctx, newMeta); err != nil {
			// 如果是重复插入（并发场景），增加引用计数
			if errors.Is(err, fileMongo.ErrAlreadyExists) {
				if err := l.svcCtx.FileMetaModel.IncrRefCount(l.ctx, in.Hash, 1); err != nil {
					l.Logger.Errorf("CreateFile IncrRefCount (concurrent) error: %v", err)
					return nil, err
				}
			} else {
				l.Logger.Errorf("CreateFile Insert file_meta error: %v", err)
				return nil, err
			}
		}
	}

	// 2. 创建 MySQL user_repository 记录
	userRepo := &model.UserRepository{
		Identity: identity,
		Hash:     in.Hash,
		UserId:   uint64(in.UserId),
		ParentId: uint64(in.ParentId),
		Name:     in.Name,
		Ext:      in.Ext,
		Size:     in.Size,
		Path:     in.S3Key,
	}

	result, err := l.svcCtx.UserRepositoryModel.Insert(l.ctx, nil, userRepo)
	if err != nil {
		l.Logger.Errorf("CreateFile Insert user_repository error: %v", err)
		// 回滚 MongoDB 引用计数
		_ = l.svcCtx.FileMetaModel.DecrRefCount(l.ctx, in.Hash, 1)
		return nil, xerr.NewErrCode(xerr.FILE_UPLOAD_FAILED)
	}

	fileId, _ := result.LastInsertId()

	// 3. 清除文件列表缓存
	if l.svcCtx.FileCache != nil {
		if err := l.svcCtx.FileCache.InvalidateUserFileListCache(l.ctx, in.UserId, in.ParentId); err != nil {
			l.Logger.Errorf("CreateFile InvalidateUserFileListCache error: %v", err)
		}
	}

	// 4. 发送 Kafka 事件
	if err := l.svcCtx.KafkaProducer.SendFileUploaded(
		l.ctx,
		in.UserId,
		fileId,
		identity,
		in.Name,
		in.Hash,
		in.Size,
		in.Ext,
	); err != nil {
		// Kafka 发送失败不影响主流程，只记录日志
		l.Logger.Errorf("CreateFile SendFileUploaded error: %v", err)
	}

	return &pb.CreateFileResp{
		Identity: identity,
		Id:       fileId,
	}, nil
}

package logic

import (
	"context"
	"errors"

	"polaris-io/backend/app/file/cmd/rpc/internal/svc"
	"polaris-io/backend/app/file/cmd/rpc/pb"
	fileMongo "polaris-io/backend/app/file/mongo"
	"polaris-io/backend/pkg/filecache"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckInstantUploadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckInstantUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckInstantUploadLogic {
	return &CheckInstantUploadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CheckInstantUpload 秒传检查
// 检查文件是否已存在于 MongoDB file_meta 集合中
// 如果存在，返回文件元数据，客户端可以直接秒传
// 使用 Redis 缓存加速查询
func (l *CheckInstantUploadLogic) CheckInstantUpload(in *pb.CheckInstantUploadReq) (*pb.CheckInstantUploadResp, error) {
	// 参数校验
	if in.Hash == "" {
		return nil, errors.New("hash is required")
	}

	// 1. 先查 Redis 缓存
	if l.svcCtx.FileCache != nil {
		cached, exists, err := l.svcCtx.FileCache.GetFileMeta(l.ctx, in.Hash)
		if err != nil {
			l.Logger.Errorf("CheckInstantUpload cache get error: %v", err)
			// 缓存出错，降级查 MongoDB
		} else if exists {
			// 缓存命中，校验大小
			l.Logger.Infof("CheckInstantUpload cache hit: hash=%s", in.Hash)
			if cached.Size != in.Size {
				l.Logger.Infof("Hash collision detected (from cache): hash=%s, expected_size=%d, actual_size=%d",
					in.Hash, in.Size, cached.Size)
				return &pb.CheckInstantUploadResp{
					Exists: false,
					Meta:   nil,
				}, nil
			}
			return &pb.CheckInstantUploadResp{
				Exists: true,
				Meta: &pb.FileMeta{
					Id:       cached.ID,
					Hash:     cached.Hash,
					Size:     cached.Size,
					S3Key:    cached.S3Key,
					Ext:      cached.Ext,
					MimeType: cached.MimeType,
					RefCount: cached.RefCount,
				},
			}, nil
		}
	}

	// 2. 缓存未命中，查 MongoDB
	l.Logger.Infof("CheckInstantUpload cache miss, querying MongoDB: hash=%s", in.Hash)
	meta, err := l.svcCtx.FileMetaModel.FindByHash(l.ctx, in.Hash)
	if err != nil {
		// 未找到，说明不能秒传
		if errors.Is(err, fileMongo.ErrNotFound) {
			return &pb.CheckInstantUploadResp{
				Exists: false,
				Meta:   nil,
			}, nil
		}
		// 其他错误
		l.Logger.Errorf("CheckInstantUpload FindByHash error: %v", err)
		return nil, err
	}

	// 找到了，校验文件大小是否一致
	if meta.Size != in.Size {
		// 大小不一致，可能是 hash 碰撞（极少见），不允许秒传
		l.Logger.Infof("Hash collision detected: hash=%s, expected_size=%d, actual_size=%d",
			in.Hash, in.Size, meta.Size)
		return &pb.CheckInstantUploadResp{
			Exists: false,
			Meta:   nil,
		}, nil
	}

	// 3. 写入缓存（24小时过期）
	if l.svcCtx.FileCache != nil {
		cacheData := &filecache.FileMetaCache{
			ID:       meta.ID.Hex(),
			Hash:     meta.Hash,
			Size:     meta.Size,
			S3Key:    meta.S3Key,
			Ext:      meta.Ext,
			MimeType: meta.MimeType,
			RefCount: meta.RefCount,
		}
		if err := l.svcCtx.FileCache.SetFileMeta(l.ctx, in.Hash, cacheData); err != nil {
			l.Logger.Errorf("CheckInstantUpload cache set error: %v", err)
			// 缓存写入失败不影响主流程
		}
	}

	return &pb.CheckInstantUploadResp{
		Exists: true,
		Meta: &pb.FileMeta{
			Id:       meta.ID.Hex(),
			Hash:     meta.Hash,
			Size:     meta.Size,
			S3Key:    meta.S3Key,
			Ext:      meta.Ext,
			MimeType: meta.MimeType,
			RefCount: meta.RefCount,
		},
	}, nil
}

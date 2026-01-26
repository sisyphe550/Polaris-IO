package logic

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"polaris-io/backend/app/file/cmd/rpc/fileservice"
	"polaris-io/backend/app/share/cmd/rpc/internal/svc"
	"polaris-io/backend/app/share/cmd/rpc/pb"
	"polaris-io/backend/app/share/model"
	"polaris-io/backend/pkg/asynqjob"
	"polaris-io/backend/pkg/xerr"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateShareLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateShareLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateShareLogic {
	return &CreateShareLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateShare 创建分享
func (l *CreateShareLogic) CreateShare(in *pb.CreateShareReq) (*pb.CreateShareResp, error) {
	// 参数校验
	if in.UserId == 0 || in.RepositoryIdentity == "" {
		return nil, errors.New("userId and repositoryIdentity are required")
	}

	// 1. 验证文件是否存在（调用 file-rpc）
	fileResp, err := l.svcCtx.FileRpc.GetFileInfo(l.ctx, &fileservice.GetFileInfoReq{
		Identity: in.RepositoryIdentity,
		UserId:   in.UserId,
	})
	if err != nil {
		l.Logger.Errorf("CreateShare GetFileInfo error: %v", err)
		return nil, xerr.NewErrCode(xerr.SHARE_FILE_NOT_EXIST)
	}
	if fileResp.File == nil {
		return nil, xerr.NewErrCode(xerr.SHARE_FILE_NOT_EXIST)
	}

	// 2. 检查是否已分享过该文件（可选：允许重复分享则注释掉这段）
	// existingShare, err := l.svcCtx.ShareModel.FindOneByRepositoryIdentity(l.ctx, uint64(in.UserId), in.RepositoryIdentity)
	// if err == nil && existingShare != nil {
	// 	return nil, xerr.NewErrCode(xerr.SHARE_ALREADY_EXISTS)
	// }

	// 3. 生成分享标识
	identity := uuid.New().String()

	// 4. 生成提取码（如果需要）
	var code string
	if in.HasCode {
		code = generateCode(l.svcCtx.Config.Share.CodeLength)
	}

	// 5. 计算过期时间
	var expiredTime uint64
	if in.ExpiredType > 0 {
		expiredTime = uint64(time.Now().Add(time.Duration(in.ExpiredType) * 24 * time.Hour).Unix())
	}

	// 6. 创建分享记录
	share := &model.Share{
		Identity:           identity,
		UserId:             uint64(in.UserId),
		RepositoryIdentity: in.RepositoryIdentity,
		Code:               code,
		ClickNum:           0,
		ExpiredTime:        expiredTime,
		Status:             0,
	}

	_, err = l.svcCtx.ShareModel.Insert(l.ctx, nil, share)
	if err != nil {
		l.Logger.Errorf("CreateShare Insert error: %v", err)
		return nil, xerr.NewErrCode(xerr.SHARE_CREATE_FAILED)
	}

	// 7. 如果设置了过期时间，入队分享过期任务
	if expiredTime > 0 && l.svcCtx.AsynqClient != nil {
		delay := time.Until(time.Unix(int64(expiredTime), 0))
		if delay > 0 {
			err = l.svcCtx.AsynqClient.EnqueueShareExpire(l.ctx, asynqjob.ShareExpirePayload{
				ShareIdentity: identity,
				UserId:        in.UserId,
			}, delay)
			if err != nil {
				// 入队失败只记录日志，不影响主流程
				l.Logger.Errorf("CreateShare EnqueueShareExpire error: %v", err)
			}
		}
	}

	return &pb.CreateShareResp{
		Identity: identity,
		Code:     code,
	}, nil
}

// generateCode 生成随机提取码
func generateCode(length int) string {
	if length <= 0 {
		length = 4
	}
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := make([]byte, length)
	for i := range code {
		code[i] = charset[r.Intn(len(charset))]
	}
	return string(code)
}

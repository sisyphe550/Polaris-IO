package logic

import (
	"context"

	"polaris-io/backend/app/file/cmd/rpc/fileservice"
	"polaris-io/backend/app/share/cmd/rpc/internal/svc"
	"polaris-io/backend/app/share/cmd/rpc/pb"
	"polaris-io/backend/app/user/cmd/rpc/usercenter"
	"polaris-io/backend/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetShareDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetShareDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetShareDetailLogic {
	return &GetShareDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetShareDetail 获取分享详情（验证提取码）
func (l *GetShareDetailLogic) GetShareDetail(in *pb.GetShareDetailReq) (*pb.GetShareDetailResp, error) {
	// 1. 验证分享是否有效
	validateLogic := NewValidateShareLogic(l.ctx, l.svcCtx)
	validateResp, err := validateLogic.ValidateShare(&pb.ValidateShareReq{
		Identity: in.Identity,
		Code:     in.Code,
	})
	if err != nil {
		return nil, err
	}

	if !validateResp.Valid {
		// 根据原因返回不同的错误码
		switch validateResp.Reason {
		case "分享不存在":
			return nil, xerr.NewErrCode(xerr.SHARE_NOT_EXIST)
		case "分享已被封禁":
			return nil, xerr.NewErrCode(xerr.SHARE_BANNED)
		case "分享已过期":
			return nil, xerr.NewErrCode(xerr.SHARE_EXPIRED)
		case "提取码错误":
			return nil, xerr.NewErrCode(xerr.SHARE_CODE_ERROR)
		default:
			return nil, xerr.NewErrCode(xerr.SHARE_NOT_EXIST)
		}
	}

	share := validateResp.Share

	// 2. 增加点击次数
	if err := l.svcCtx.ShareModel.IncrClickNum(l.ctx, in.Identity); err != nil {
		l.Logger.Errorf("GetShareDetail IncrClickNum error: %v", err)
		// 不影响主流程
	}

	// 3. 获取分享者信息
	var sharerName string
	userResp, err := l.svcCtx.UsercenterRpc.GetUserInfo(l.ctx, &usercenter.GetUserInfoReq{
		UserId: share.UserId,
	})
	if err == nil && userResp.User != nil {
		sharerName = userResp.User.Name
		if sharerName == "" && len(userResp.User.Mobile) >= 11 {
			sharerName = userResp.User.Mobile[:3] + "****" + userResp.User.Mobile[7:]
		}
	}

	// 4. 获取文件信息（调用 file-rpc，不传 userId 跳过权限检查）
	fileResp, err := l.svcCtx.FileRpc.GetFileInfo(l.ctx, &fileservice.GetFileInfoReq{
		Identity: share.RepositoryIdentity,
		UserId:   0, // 不验证用户权限
	})
	if err != nil {
		l.Logger.Errorf("GetShareDetail GetFileInfo error: %v", err)
		return nil, xerr.NewErrCode(xerr.SHARE_FILE_NOT_EXIST)
	}

	file := fileResp.File
	shareFileInfo := &pb.ShareFileInfo{
		Identity: file.Identity,
		Name:     file.Name,
		Ext:      file.Ext,
		Size:     file.Size,
		IsDir:    file.IsDir,
		Hash:     file.Hash,
		ParentId: file.ParentId,
	}

	// 5. 如果是文件夹，获取子文件列表
	var children []*pb.ShareFileInfo
	if file.IsDir {
		listResp, err := l.svcCtx.FileRpc.ListFiles(l.ctx, &fileservice.ListFilesReq{
			UserId:   share.UserId,
			ParentId: file.Id,
			Page:     1,
			PageSize: 100, // 最多返回 100 个子文件
		})
		if err == nil && listResp.List != nil {
			children = make([]*pb.ShareFileInfo, 0, len(listResp.List))
			for _, f := range listResp.List {
				children = append(children, &pb.ShareFileInfo{
					Identity: f.Identity,
					Name:     f.Name,
					Ext:      f.Ext,
					Size:     f.Size,
					IsDir:    f.IsDir,
					Hash:     f.Hash,
					ParentId: f.ParentId,
				})
			}
		}
	}

	return &pb.GetShareDetailResp{
		Share:      share,
		File:       shareFileInfo,
		SharerName: sharerName,
		Children:   children,
	}, nil
}

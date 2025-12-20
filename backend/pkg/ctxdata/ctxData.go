package ctxdata

import (
	"context"
	"encoding/json"

	"github.com/zeromicro/go-zero/core/logx"
)

// CtxKeyJwtUserId get uid from ctx
// 定义为常量，确保该 Key 与你在 Token 生成逻辑中写入 Claims 的 Key 一致
const CtxKeyJwtUserId = "jwtUserId"

// GetUidFromCtx 从 Context 中获取用户 ID
func GetUidFromCtx(ctx context.Context) int64 {
	var uid int64
	// 从 Context 中取出由 jwt 中间件注入的值
	if jsonUid, ok := ctx.Value(CtxKeyJwtUserId).(json.Number); ok {
		if int64Uid, err := jsonUid.Int64(); err == nil {
			uid = int64Uid
		} else {
			logx.WithContext(ctx).Errorf("GetUidFromCtx err : %+v", err)
		}
	}
	return uid
}

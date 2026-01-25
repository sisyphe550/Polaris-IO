package access

import (
	"net/http"

	"polaris-io/backend/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
	"polaris-io/backend/app/share/cmd/api/internal/logic/access"
	"polaris-io/backend/app/share/cmd/api/internal/svc"
	"polaris-io/backend/app/share/cmd/api/internal/types"
)

// 获取分享详情（验证提取码）
func ShareDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ShareDetailReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := access.NewShareDetailLogic(r.Context(), svcCtx)
		resp, err := l.ShareDetail(&req)
		result.HttpResult(r, w, resp, err)
	}
}

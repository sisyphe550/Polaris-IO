package share

import (
	"net/http"

	"polaris-io/backend/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
	"polaris-io/backend/app/share/cmd/api/internal/logic/share"
	"polaris-io/backend/app/share/cmd/api/internal/svc"
	"polaris-io/backend/app/share/cmd/api/internal/types"
)

// 取消分享
func CancelShareHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CancelShareReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := share.NewCancelShareLogic(r.Context(), svcCtx)
		resp, err := l.CancelShare(&req)
		result.HttpResult(r, w, resp, err)
	}
}

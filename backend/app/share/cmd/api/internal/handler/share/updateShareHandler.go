package share

import (
	"net/http"

	"polaris-io/backend/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
	"polaris-io/backend/app/share/cmd/api/internal/logic/share"
	"polaris-io/backend/app/share/cmd/api/internal/svc"
	"polaris-io/backend/app/share/cmd/api/internal/types"
)

// 更新分享
func UpdateShareHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateShareReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := share.NewUpdateShareLogic(r.Context(), svcCtx)
		resp, err := l.UpdateShare(&req)
		result.HttpResult(r, w, resp, err)
	}
}

package save

import (
	"net/http"

	"polaris-io/backend/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
	"polaris-io/backend/app/share/cmd/api/internal/logic/save"
	"polaris-io/backend/app/share/cmd/api/internal/svc"
	"polaris-io/backend/app/share/cmd/api/internal/types"
)

// 保存分享到我的网盘
func SaveShareHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SaveShareReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := save.NewSaveShareLogic(r.Context(), svcCtx)
		resp, err := l.SaveShare(&req)
		result.HttpResult(r, w, resp, err)
	}
}

package search

import (
	"net/http"

	"polaris-io/backend/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
	"polaris-io/backend/app/search/cmd/api/internal/logic/search"
	"polaris-io/backend/app/search/cmd/api/internal/svc"
	"polaris-io/backend/app/search/cmd/api/internal/types"
)

// 获取用户文件统计
func GetUserStatsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetUserStatsReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := search.NewGetUserStatsLogic(r.Context(), svcCtx)
		resp, err := l.GetUserStats(&req)
		result.HttpResult(r, w, resp, err)
	}
}

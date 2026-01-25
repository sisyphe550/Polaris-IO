package search

import (
	"net/http"

	"polaris-io/backend/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
	"polaris-io/backend/app/search/cmd/api/internal/logic/search"
	"polaris-io/backend/app/search/cmd/api/internal/svc"
	"polaris-io/backend/app/search/cmd/api/internal/types"
)

// 搜索文件
func SearchFilesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SearchFilesReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := search.NewSearchFilesLogic(r.Context(), svcCtx)
		resp, err := l.SearchFiles(&req)
		result.HttpResult(r, w, resp, err)
	}
}

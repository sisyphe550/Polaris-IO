package access

import (
	"net/http"

	"polaris-io/backend/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
	"polaris-io/backend/app/share/cmd/api/internal/logic/access"
	"polaris-io/backend/app/share/cmd/api/internal/svc"
	"polaris-io/backend/app/share/cmd/api/internal/types"
)

// 获取分享文件下载链接
func ShareDownloadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ShareDownloadReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := access.NewShareDownloadLogic(r.Context(), svcCtx)
		resp, err := l.ShareDownload(&req)
		result.HttpResult(r, w, resp, err)
	}
}

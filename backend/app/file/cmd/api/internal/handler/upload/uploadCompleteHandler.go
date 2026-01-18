package upload

import (
	"net/http"

	"polaris-io/backend/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
	"polaris-io/backend/app/file/cmd/api/internal/logic/upload"
	"polaris-io/backend/app/file/cmd/api/internal/svc"
	"polaris-io/backend/app/file/cmd/api/internal/types"
)

// 上传完成回调
func UploadCompleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UploadCompleteReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := upload.NewUploadCompleteLogic(r.Context(), svcCtx)
		resp, err := l.UploadComplete(&req)
		result.HttpResult(r, w, resp, err)
	}
}

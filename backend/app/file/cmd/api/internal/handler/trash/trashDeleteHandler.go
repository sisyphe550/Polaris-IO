package trash

import (
	"net/http"

	"polaris-io/backend/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
	"polaris-io/backend/app/file/cmd/api/internal/logic/trash"
	"polaris-io/backend/app/file/cmd/api/internal/svc"
	"polaris-io/backend/app/file/cmd/api/internal/types"
)

// 彻底删除
func TrashDeleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TrashDeleteReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := trash.NewTrashDeleteLogic(r.Context(), svcCtx)
		resp, err := l.TrashDelete(&req)
		result.HttpResult(r, w, resp, err)
	}
}

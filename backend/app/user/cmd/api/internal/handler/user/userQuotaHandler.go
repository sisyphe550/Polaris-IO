package user

import (
	"net/http"

	"polaris-io/backend/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
	"polaris-io/backend/app/user/cmd/api/internal/logic/user"
	"polaris-io/backend/app/user/cmd/api/internal/svc"
	"polaris-io/backend/app/user/cmd/api/internal/types"
)

// 获取当前用户配额
func UserQuotaHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserQuotaReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := user.NewUserQuotaLogic(r.Context(), svcCtx)
		resp, err := l.UserQuota(&req)
		result.HttpResult(r, w, resp, err)
	}
}

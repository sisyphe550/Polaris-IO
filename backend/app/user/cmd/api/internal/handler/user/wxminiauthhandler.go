package user

import (
	"net/http"

	"shared-board/backend/pkg/result"

	"shared-board/backend/app/user/cmd/api/internal/logic/user"
	"shared-board/backend/app/user/cmd/api/internal/svc"
	"shared-board/backend/app/user/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 微信小程序授权登录(通常需要Token关联)
func WxMiniAuthHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.WxMiniAuthReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := user.NewWxMiniAuthLogic(r.Context(), svcCtx)
		resp, err := l.WxMiniAuth(&req)
		result.HttpResult(r, w, resp, err)
	}
}

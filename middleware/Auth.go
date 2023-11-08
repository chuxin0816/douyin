package middleware

import (
	"context"
	"douyin/pkg/jwt"
	"douyin/response"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

const CtxUserIDKey = "userID"

func AuthMiddleware() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		// 获取token
		token, ok := ctx.GetQuery("token")
		if !ok {
			token, ok = ctx.GetPostForm("token")
			if !ok {
				response.Error(ctx, response.CodeInvalidParam)
				hlog.Error("AuthMiddleware: token不存在")
				ctx.Abort()
				return
			}
		}

		// 验证token
		if len(token) == 0 {
			ctx.Abort()
		}
		userID, err := jwt.ParseToken(token)
		if err != nil {
			response.Error(ctx, response.CodeNoAuthority)
			hlog.Error("AuthMiddleware: token无效, err: ", err)
			ctx.Abort()
			return
		}

		// 设置userID到上下文
		ctx.Set(CtxUserIDKey, userID)
		ctx.Next(c)
	}
}

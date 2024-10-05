package mw

import (
	"context"

	"douyin/src/common/jwt"
	"douyin/src/service/api/controller"

	"github.com/cloudwego/hertz/pkg/app"
)

func AuthMiddleware() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		// 获取token
		token, ok := ctx.GetQuery("token")
		if !ok {
			token, ok = ctx.GetPostForm("token")
			if !ok {
				controller.Error(ctx, controller.CodeInvalidParam)
				ctx.Abort()
				return
			}
		}

		// 验证token
		if len(token) == 0 {
			controller.Error(ctx, controller.CodeInvalidParam)
			ctx.Abort()
		}
		userID := jwt.ParseToken(token)
		if userID == nil {
			controller.Error(ctx, controller.CodeNoAuthority)
			ctx.Abort()
			return
		}

		// 设置userID到上下文
		ctx.Set(controller.CtxUserIDKey, *userID)
		ctx.Next(c)
	}
}

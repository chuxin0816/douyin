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
			return
		}

		userID := jwt.ParseAccessToken(token)
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

func RefreshTokenMiddleware() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		// 获取token
		token, ok := ctx.GetQuery("refresh_token")
		if !ok {
			token, ok = ctx.GetPostForm("refresh_token")
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
			return
		}

		userID := jwt.ParseRefreshToken(token)
		if userID == nil {
			controller.Error(ctx, controller.CodeNoAuthority)
			ctx.Abort()
			return
		}

		newToken, err := jwt.GenerateAccessToken(*userID)
		if err != nil {
			controller.Error(ctx, controller.CodeServerBusy)
			ctx.Abort()
			return
		}

		// 返回新的token
		controller.Success(ctx, RefreshTokenResponse{
			Token: newToken,
		})
	}
}

type RefreshTokenResponse struct {
	StatusCode int32  `json:"status_code"`
	Token      string `json:"token,omitempty"`
}

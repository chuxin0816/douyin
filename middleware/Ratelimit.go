package middleware

import (
	"context"
	"douyin/controller"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/juju/ratelimit"
)

func RatelimitMiddleware(fillInterval time.Duration, capacity int64) app.HandlerFunc {
	rl := ratelimit.NewBucket(fillInterval, capacity)
	return func(c context.Context, ctx *app.RequestContext) {
		if rl.TakeAvailable(1) < 1 {
			controller.Error(ctx, controller.CodeServerBusy)
			ctx.Abort()
			return
		}
		ctx.Next(c)
	}
}

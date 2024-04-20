package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/juju/ratelimit"
)

// RatelimitMiddleware 令牌桶限流中间件，fillInterval为填充间隔/牌，capacity为桶容量
func RatelimitMiddleware(QPS int64) app.HandlerFunc {
	rl := ratelimit.NewBucket(time.Second/time.Duration(QPS), QPS)
	return func(c context.Context, ctx *app.RequestContext) {
		if rl.TakeAvailable(1) == 0 {
			ctx.JSON(consts.StatusTooManyRequests, nil)
			ctx.Abort()
			return
		}
		ctx.Next(c)
	}
}

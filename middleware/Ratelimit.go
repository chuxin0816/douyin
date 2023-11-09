package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/juju/ratelimit"
)

func RatelimitMiddleware(fillInterval time.Duration, capacity int64) app.HandlerFunc {
	rl := ratelimit.NewBucket(fillInterval, capacity)
	return func(c context.Context, ctx *app.RequestContext) {
		if rl.TakeAvailable(1) < 1 {
			ctx.String(consts.StatusOK, "rate limit...")
			ctx.Abort()
			return
		}
		ctx.Next(c)
	}
}

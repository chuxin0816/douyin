package middleware

import (
	"context"
	"douyin/src/dal"

	"github.com/chuxin0816/ratelimit"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

const bucketName = "ratelimit"

// RatelimitMiddleware 令牌桶限流中间件，fillInterval为填充间隔/牌，capacity为桶容量
func RatelimitMiddleware(rate, capacity int) app.HandlerFunc {
	bucket := ratelimit.NewBucket(dal.RDB, dal.GetRedisKey(bucketName), rate, capacity)
	return func(c context.Context, ctx *app.RequestContext) {
		if !bucket.Take() {
			ctx.JSON(consts.StatusTooManyRequests, nil)
			ctx.Abort()
			return
		}
		ctx.Next(c)
	}
}

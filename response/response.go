package response

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func Success(ctx *app.RequestContext, data any) {
	ctx.JSON(consts.StatusOK, data)
}

func Error(ctx *app.RequestContext, code ResCode) {
	ctx.JSON(consts.StatusOK, &Response{StatusCode: code, StatusMsg: code.Msg()})
}

type Response struct {
	StatusCode ResCode `json:"status_code"` // 状态码，0-成功，其他值-失败
	StatusMsg  string  `json:"status_msg"`  // 返回状态描述
}

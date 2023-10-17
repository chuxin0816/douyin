package response

type ResCode int32

const (
	CodeSuccess     ResCode = 0
	CodeNoAuthority ResCode = 1000 + iota
	CodeInvalidParam
	CodeUserExist
	CodeUserNotExist
	CodeInvalidPassword
	CodeFileTooLarge
	CodeAlreadyFavorite
	CodeNotFavorite
	CodeServerBusy
)

var codeMsgMap = map[ResCode]string{
	CodeSuccess:         "请求成功",
	CodeNoAuthority:     "权限不足",
	CodeInvalidParam:    "请求参数错误",
	CodeUserExist:       "用户名已存在",
	CodeUserNotExist:    "用户名不存在",
	CodeInvalidPassword: "密码错误",
	CodeFileTooLarge:    "文件太大",
	CodeAlreadyFavorite: "已经点赞过了",
	CodeNotFavorite:     "还没有点赞过",
	CodeServerBusy:      "服务器繁忙",
}

func (code ResCode) Msg() string {
	msg, ok := codeMsgMap[code]
	if !ok {
		msg = codeMsgMap[CodeServerBusy]
	}
	return msg
}

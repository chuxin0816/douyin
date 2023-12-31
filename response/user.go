package response

// UserResponse 用户响应
type UserResponse struct {
	ID              int64  `json:"id"`               // 用户id
	Name            string `json:"name"`             // 用户名称
	Avatar          string `json:"avatar"`           // 用户头像
	BackgroundImage string `json:"background_image"` // 用户个人页顶部大图
	FavoriteCount   int64  `json:"favorite_count"`   // 喜欢数
	FollowCount     int64  `json:"follow_count"`     // 关注总数
	FollowerCount   int64  `json:"follower_count"`   // 粉丝总数
	WorkCount       int64  `json:"work_count"`       // 作品数
	IsFollow        bool   `json:"is_follow"`        // true-已关注，false-未关注
	Signature       string `json:"signature"`        // 个人简介
	TotalFavorited  string `json:"total_favorited"`  // 获赞数量
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	*Response
	UserID int64  `json:"user_id"` // 用户id
	Token  string `json:"token"`   // 用户鉴权token
}

// LoginResponse 登录响应
type LoginResponse struct {
	*Response
	UserID int64  `json:"user_id"` // 用户id
	Token  string `json:"token"`   // 用户鉴权token
}

type UserInfoResponse struct {
	*Response
	User *UserResponse `json:"user"` // 用户信息
}

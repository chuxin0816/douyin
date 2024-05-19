namespace go user

struct User {
  1: i64 id; // 用户id
  2: string name; // 用户名称
  3: i64 follow_count; // 关注总数
  4: i64 follower_count; // 粉丝总数
  5: bool is_follow; // true-已关注，false-未关注
  6: string avatar; //用户头像
  7: string background_image; //用户个人页顶部大图
  8: string signature; //个人简介
  9: i64 total_favorited; //获赞数量
  10: i64 work_count; //作品数量
  11: i64 favorite_count; //点赞数量
}

struct User_register_request {
  1: string username; // 注册用户名，最长32个字符
  2: string password; // 密码，最长32个字符
}

struct User_register_response {
  1: i32 status_code; // 状态码，0-成功，其他值-失败 404-用户名已存在
  2: optional string status_msg; // 返回状态描述
  3: i64 user_id; // 用户id
  4: string token; // 用户鉴权token
}

struct User_login_request {
  1: string username; // 登录用户名
  2: string password; // 登录密码
}

struct User_login_response {
  1: i32 status_code; // 状态码，0-成功，其他值-失败
  2: optional string status_msg; // 返回状态描述
  3: i64 user_id; // 用户id
  4: string token; // 用户鉴权token
}

struct User_info_request {
  1: optional i64 user_id; // 用户id
  2: i64 author_id; // 对方用户id
}

struct User_info_response {
  1: i32 status_code; // 状态码，0-成功，其他值-失败
  2: optional string status_msg; // 返回状态描述
  3: User user; // 用户信息
}

service UserService {
    User_register_response Register(1: User_register_request req);
    User_login_response Login(1: User_login_request req);
    User_info_response UserInfo(1: User_info_request req);
}
include "feed.thrift"

namespace go favorite

struct Favorite_action_request {
  1: i64 user_id; // 用户id
  2: i64 video_id; // 视频id
  3: i64 action_type; // 1-点赞，2-取消点赞
}

struct Favorite_action_response {
  1: i32 status_code; // 状态码，0-成功，其他值-失败
  2: optional string status_msg; // 返回状态描述
}

struct Favorite_list_request {
  1: optional i64 user_id; // 用户id
  2: i64 to_user_id; // 对方用户id
}

struct Favorite_list_response {
  1: i32 status_code; // 状态码，0-成功，其他值-失败
  2: optional string status_msg; // 返回状态描述
  3: list<feed.Video> video_list; // 用户点赞视频列表
}

service FavoriteService {
    Favorite_action_response FavoriteAction(1: Favorite_action_request req);
    Favorite_list_response FavoriteList(1: Favorite_list_request req);
}


include "feed.thrift"

namespace go publish

struct Publish_action_request {
  1: i64 user_id; // 用户id
  2: binary data; // 视频数据
  3: string title; // 视频标题
}

struct Publish_action_response {
  1: i32 status_code; // 状态码，0-成功，其他值-失败
  2: optional string status_msg; // 返回状态描述
}

struct Publish_list_request {
  1: i64 user_id; // 用户id
  2: i64 to_user_id; // 对方用户id
}

struct Publish_list_response {
  1: i32 status_code; // 状态码，0-成功，其他值-失败
  2: optional string status_msg; // 返回状态描述
  3: list<feed.Video> video_list; // 用户发布的视频列表
}

service PublishService {
    Publish_action_response PublishAction(1: Publish_action_request req)
    Publish_list_response PublishList(1: Publish_list_request req)
}
include "user.thrift"

namespace go video

struct Video {
  1:  i64 id; // 视频唯一标识
  2:  user.User author; // 视频作者信息
  3:  string play_url; // 视频播放地址
  4:  string cover_url; // 视频封面地址
  5:  i64 upload_time; // 视频上传时间，精确到秒
  6:  i64 favorite_count; // 视频的点赞总数
  7:  i64 comment_count; // 视频的评论总数
  8:  bool is_favorite; // true-已点赞，false-未点赞
  9:  string title; // 视频标题
}

struct Feed_request {
  1: i64 latest_time; // 可选参数，限制返回视频的最新投稿时间戳，精确到秒，不填表示当前时间
  2: optional i64 user_id; // 用户id
}

struct Feed_response {
  1: i32 status_code; // 状态码，0-成功，其他值-失败
  2: optional string status_msg; // 返回状态描述
  3: list<Video> video_list; // 视频列表
  4: optional i64 next_time; // 本次返回的视频中，发布最早的时间，作为下次请求时的latest_time
}

struct Video_info_request {
  1: optional i64 user_id; // 用户id
  2: i64 video_id; // 视频id
}

struct Video_info_list_request {
  1: optional i64 user_id; // 用户id
  2: list<i64> video_id_list; // 视频id列表
}

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
  1: optional i64 user_id; // 用户id
  2: i64 author_id; // 对方用户id
}

struct Publish_list_response {
  1: i32 status_code; // 状态码，0-成功，其他值-失败
  2: optional string status_msg; // 返回状态描述
  3: list<Video> video_list; // 用户发布的视频列表
}

service VideoService {
  Feed_response Feed(1: Feed_request req);
  Publish_action_response PublishAction(1: Publish_action_request req)
  Publish_list_response PublishList(1: Publish_list_request req)
  list<i64> PublishIDList(1: i64 user_id)
  Video VideoInfo(1: Video_info_request req);
  list<Video> VideoInfoList(1: Video_info_list_request req);
  i64 WorkCount(1: i64 user_id)
  i64 AuthorId(1: i64 video_id)
  bool VideoExist(1: i64 video_id)
}
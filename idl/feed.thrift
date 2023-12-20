include "user.thrift"

namespace go feed

struct Feed_request {
  1: optional i64 latest_time; // 可选参数，限制返回视频的最新投稿时间戳，精确到秒，不填表示当前时间
  2: optional string token; // 可选参数，登录用户设置
}

struct Feed_response {
  1: i32 status_code; // 状态码，0-成功，其他值-失败
  2: optional string status_msg; // 返回状态描述
  3: list<Video> video_list; // 视频列表
  4: optional i64 next_time; // 本次返回的视频中，发布最早的时间，作为下次请求时的latest_time
}

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

service FeedService {
  douyin_feed_response Feed(1: douyin_feed_request req);
}





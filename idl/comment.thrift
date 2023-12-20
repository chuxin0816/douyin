include "user.thrift"

namespace go comment

struct Comment_action_request {
  1: string token; // 用户鉴权token
  2: i64 video_id; // 视频id
  3: i32 action_type; // 1-发布评论，2-删除评论
  4: optional string comment_text; // 用户填写的评论内容，在action_type=1的时候使用
  5: optional string comment_id; // 要删除的评论id，在action_type=2的时候使用
}

struct Comment_action_response {
  1: i32 status_code; // 状态码，0-成功，其他值-失败
  2: optional string status_msg; // 返回状态描述
  3: optional Comment comment; // 评论成功返回评论内容，不需要重新拉取整个列表
}

struct Comment_list_request {
  1: string token; // 用户鉴权token
  2: i64 video_id; // 视频id
}

struct Comment_list_response {
  1: i32 status_code; // 状态码，0-成功，其他值-失败
  2: optional string status_msg; // 返回状态描述
  3: list<Comment> comment_list; // 评论列表
}

struct Comment {
  1: i64 id; // 视频评论id
  2: user.User user; // 评论用户信息
  3: string content; // 评论内容
  4: string create_date; // 评论发布日期，格式 mm-dd
}

service CommentService {
    Comment_action_response CommentAction(1:Comment_action_request req)
    Comment_list_response CommentList(1:Comment_list_request req)
}
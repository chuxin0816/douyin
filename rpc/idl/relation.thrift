include "user.thrift"

namespace go relation

struct Relation_action_request {
  1: i64 user_id; // 用户id
  2: i64 to_user_id; // 对方用户id
  3: i64 action_type; // 1-关注，2-取消关注
}

struct Relation_action_response {
  1: i32 status_code; // 状态码，0-成功，其他值-失败
  2: optional string status_msg; // 返回状态描述
}

struct Relation_follow_list_request {
  1: i64 to_user_id; // 对方用户id
  2: i64 user_id; // 用户id
}

struct Relation_follow_list_response {
  1: i32 status_code; // 状态码，0-成功，其他值-失败
  2: optional string status_msg; // 返回状态描述
  3: list<user.User> user_list; // 用户信息列表
}

struct Relation_follower_list_request {
  1: i64 to_user_id; // 对方用户id
  2: i64 user_id; // 用户id
}

struct Relation_follower_list_response {
  1: i32 status_code; // 状态码，0-成功，其他值-失败
  2: optional string status_msg; // 返回状态描述
  3: list<user.User> user_list; // 用户列表
}

struct Relation_friend_list_request {
  1: i64 to_user_id; // 对方用户id
  2: i64 user_id; // 用户id
}

struct Relation_friend_list_response {
  1: i32 status_code ; // 状态码，0-成功，其他值-失败
  2: optional string status_msg; // 返回状态描述
  3: list<user.User> user_list; // 用户列表
}

service RelationService{
    Relation_action_response RelationAction(1: Relation_action_request req)
    Relation_follow_list_response RelationFollowList(1: Relation_follow_list_request req)
    Relation_follower_list_response RelationFollowerList(1: Relation_follower_list_request req)
    Relation_friend_list_response RelationFriendList(1: Relation_friend_list_request req)
}
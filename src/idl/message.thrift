namespace go message

struct Message {
  1: i64 id; // 消息id
  2: i64 to_user_id; // 该消息接收者的id
  3: i64 from_user_id; // 该消息发送者的id
  4: string content; // 消息内容
  5: i64 create_time; // 消息发送时间，时间戳
}

struct Message_action_request {
  1: i64 user_id; // 用户id
  2: i64 to_user_id; // 对方用户id
  3: i64 action_type; // 1-发送消息
  4: string content; // 消息内容
}

struct Message_action_response {
  1: i32 status_code; // 状态码，0-成功，其他值-失败
  2: optional string status_msg; // 返回状态描述
}

struct Message_chat_request {
  1: i64 user_id; // 用户id
  2: i64 to_user_id; // 对方用户id
  3: i64 last_time;//上次最新消息的时间（新增字段-apk更新中）
}

struct Message_chat_response {
  1: i32 status_code; // 状态码，0-成功，其他值-失败
  2: optional string status_msg; // 返回状态描述
  3: list<Message> message_list; // 消息列表
}

service MessageService{
    Message_chat_response MessageChat(1: Message_chat_request req)
    Message_action_response MessageAction(1: Message_action_request req)
}
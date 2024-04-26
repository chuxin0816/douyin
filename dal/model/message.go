package model

const TableNameMessage = "message"

type Message struct {
	ID         int64  `json:"id" bson:"_id"`
	ToUserID   int64  `json:"to_user_id" bson:"to_user_id"`
	FromUserID int64  `json:"from_user_id" bson:"from_user_id"`
	ConvertID  string  `json:"convert_id" bson:"convert_id"`
	Content    string `json:"content" bson:"content"`
	CreateTime int64  `json:"create_time" bson:"create_time"`
}

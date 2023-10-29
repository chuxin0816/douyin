package models

import "time"

type Message struct {
	ID         int64     `json:"id"           gorm:"primaryKey"`        // 消息id
	ToUserID   int64     `json:"to_user_id"   gorm:"index:idx_from_to"` // 消息接收者id
	FromUserID int64     `json:"from_user_id" gorm:"index:idx_from_to"` // 消息发送者id
	Content    string    `json:"content"      gorm:"type:varchar(255)"` // 消息内容
	CreateTime time.Time `json:"create_time"`                           // 消息发送时间
}

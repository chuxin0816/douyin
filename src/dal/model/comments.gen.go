// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"
)

const TableNameComment = "comments"

// Comment mapped from table <comments>
type Comment struct {
	ID         int64     `gorm:"column:id;primaryKey" json:"id"`
	VideoID    int64     `gorm:"column:video_id;not null;comment:视频ID" json:"video_id"`                                 // 视频ID
	UserID     int64     `gorm:"column:user_id;not null;comment:用户ID" json:"user_id"`                                   // 用户ID
	Content    string    `gorm:"column:content;not null;comment:评论内容" json:"content"`                                   // 评论内容
	CreateTime time.Time `gorm:"column:create_time;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"create_time"` // 创建时间
}

// TableName Comment's table name
func (*Comment) TableName() string {
	return TableNameComment
}
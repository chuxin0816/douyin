package models

import "time"

type Comment struct {
	ID         int64     `json:"id"       gorm:"primaryKey"`         // 评论id
	VideoID    int64     `json:"video_id" gorm:"index:idx_video_id"` // 视频id
	UserID     int64     `json:"user_id"  gorm:"index:idx_user_id"`  // 用户id
	Content    string    `json:"content"  gorm:"type:varchar(255)"`  // 评论内容
	CreateTime time.Time `json:"create_time"`                        // 创建时间
}

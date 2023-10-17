package models

import "time"

type Video struct {
	ID            int64     `json:"id" gorm:"primaryKey"`                // 视频唯一标识
	AuthorID      int64     `json:"author_id" gorm:"index"`              // 视频作者信息
	PlayURL       string    `json:"play_url"`                            // 视频播放地址
	CoverURL      string    `json:"cover_url"`                           // 视频封面地址
	UploadTime    time.Time `json:"upload_time" gorm:"index:,sort:desc"` // 视频上传时间
	FavoriteCount int64     `json:"favorite_count" gorm:"default:0"`     // 视频的点赞总数
	Title         string    `json:"title"`                               // 视频标题
	CommentCount  int64     `json:"comment_count" gorm:"default:0"`      // 视频的评论总数
}

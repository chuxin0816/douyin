package model

type Relation struct {
	ID         int64 `json:"id"`
	AuthorID   int64 `json:"author_id"`   // 作者ID
	FollowerID int64 `json:"follower_id"` // 粉丝ID
}

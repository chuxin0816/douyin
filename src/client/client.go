package client

import (
	"sync"

	"douyin/src/kitex_gen/comment/commentservice"
	"douyin/src/kitex_gen/favorite/favoriteservice"
	"douyin/src/kitex_gen/message/messageservice"
	"douyin/src/kitex_gen/relation/relationservice"
	"douyin/src/kitex_gen/user/userservice"
	"douyin/src/kitex_gen/video/videoservice"
)

var (
	CommentClient  commentservice.Client
	FavoriteClient favoriteservice.Client
	MessageClient  messageservice.Client
	RelationClient relationservice.Client
	UserClient     userservice.Client
	VideoClient    videoservice.Client
	once           sync.Once
	err            error
)

func Init() {
	once.Do(func() {
		initCommentClient()
		initFavoriteClient()
		initMessageClient()
		initRelationClient()
		initUserClient()
		initVideoClient()
	})
}

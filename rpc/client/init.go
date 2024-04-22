package client

func InitRpcClient() {
	initFeedClient()     // 8890
	initUserClient()     // 8891
	initFavoriteClient() // 8892
	initCommentClient()  // 8893
	initPublishClient()  // 8894
	initRelationClient() // 8895
	initMessageClient()  // 8896
}

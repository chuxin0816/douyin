package response

type RelationActionResponse struct {
	*Response
}

type FollowListResponse struct {
	*Response
	UserList []*UserResponse `json:"user_list"`
}

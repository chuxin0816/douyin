package response

type RelationActionResponse struct {
	*Response
}

type RelationListResponse struct {
	*Response
	UserList []*UserResponse `json:"user_list"`
}
package main

import (
	"context"
	relation "douyin/rpc/kitex_gen/relation"
)

// RelationServiceImpl implements the last service interface defined in the IDL.
type RelationServiceImpl struct{}

// RelationAction implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) RelationAction(ctx context.Context, req *relation.RelationActionRequest) (resp *relation.RelationActionResponse, err error) {
	// TODO: Your code here...
	return
}

// RelationFollowList implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) RelationFollowList(ctx context.Context, req *relation.RelationFollowListRequest) (resp *relation.RelationFollowListResponse, err error) {
	// TODO: Your code here...
	return
}

// RelationFollowerList implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) RelationFollowerList(ctx context.Context, req *relation.RelationFollowerListRequest) (resp *relation.RelationFollowerListResponse, err error) {
	// TODO: Your code here...
	return
}

// RelationFriendList implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) RelationFriendList(ctx context.Context, req *relation.RelationFriendListRequest) (resp *relation.RelationFriendListResponse, err error) {
	// TODO: Your code here...
	return
}

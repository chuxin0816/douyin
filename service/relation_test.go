package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRelationAction(t *testing.T) {
	resp, err := RelationAction(12182603931062272, 10760536648060928, 1)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, 0, int(resp.StatusCode))
	resp, err = RelationAction(12182603931062272, 10760536648060928, 2)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, 0, int(resp.StatusCode))
}

func TestFollowList(t *testing.T) {
	resp, err := FollowList(12182603931062272, 10760536648060928)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, 0, int(resp.StatusCode))
}

func TestFollowerList(t *testing.T) {
	resp, err := FollowerList(12182603931062272, 10760536648060928)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, 0, int(resp.StatusCode))
}

func TestFriendList(t *testing.T) {
	resp, err := FriendList(12182603931062272, 10760536648060928)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, 0, int(resp.StatusCode))
}

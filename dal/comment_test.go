package dal

import (
	"douyin/config"
	"douyin/pkg/snowflake"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	config.Init()
	Init()
	snowflake.Init()
}
func TestCommentAction(t *testing.T) {
	err := PublishComment(12182603931062272, 1111111111, 10760595804524544, "test comment")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	err = DeleteComment(1111111111, 10760595804524544)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestGetCommentByID(t *testing.T) {
	comment, err := GetCommentByID(12564221107638272)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, comment.ID, int64(12564221107638272))
}

func TestGetCommentList(t *testing.T) {
	comments, err := GetCommentList(10760595804524544)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, 1, len(comments))
}

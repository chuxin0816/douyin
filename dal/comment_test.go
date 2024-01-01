package dal

import (
	"context"
	"douyin/config"
	"douyin/dal/model"
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
	err := CreateComment(context.Background(), &model.Comment{
		ID:      1111111111,
		VideoID: 10760595804524544,
		UserID:  10760595804524544,
		Content: "test",
	})
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	err = DeleteComment(context.Background(), 1111111111)
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

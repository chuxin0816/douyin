package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommentAction(t *testing.T) {
	resp, err := CommentAction(12182603931062272, 1, 10760595804524544, 0, "hello")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, 0, int(resp.StatusCode))
	id := resp.Comment.ID
	resp, err = CommentAction(12182603931062272, 2, 10760595804524544, id, "")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, 0, int(resp.StatusCode))
}

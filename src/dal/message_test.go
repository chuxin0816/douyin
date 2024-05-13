package dal

import (
	"context"
	"douyin/dal/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageAction(t *testing.T) {
	err := MessageAction(context.Background(), &model.Message{})
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestMessageList(t *testing.T) {
	messages, err := MessageList(context.Background(), 12182603931062272, 10760536648060928, 0)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.NotNil(t, messages)
}

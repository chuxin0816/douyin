package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageAction(t *testing.T) {
	resp, err := MessageAction(12182603931062272, 10760536648060928, 1, "hello")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, 0, int(resp.StatusCode))
}

func TestMessageChat(t *testing.T) {
	resp, err := MessageChat(12182603931062272, 10760536648060928, 0)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, 0, int(resp.StatusCode))
}

package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublishList(t *testing.T) {
	resp, err := PublishList(12182603931062272, 10760536648060928)
	if err != nil {
		t.Log(err)
	}
	assert.Equal(t,0, int(resp.StatusCode))
}

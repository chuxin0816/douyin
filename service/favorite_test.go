package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFavoriteAction(t *testing.T) {
	resp, err := FavoriteAction(12182603931062272, 10760595804524544, 1)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, 0, int(resp.StatusCode))
	resp, err = FavoriteAction(12182603931062272, 10760595804524544, 2)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, 0, int(resp.StatusCode))
}

func TestFavoriteList(t *testing.T) {
	resp,err:=FavoriteList(12182603931062272, 10760536648060928)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, 0, int(resp.StatusCode))
}
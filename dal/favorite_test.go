package dal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFavoriteAction(t *testing.T) {
	err := FavoriteAction(12182603931062272, 10760595804524544, 1)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	err = FavoriteAction(12182603931062272, 10760595804524544, -1)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestErrVideoNotExist(t *testing.T) {
	err := FavoriteAction(12182603931062272, 11111111111111, -1)
	assert.Equal(t, ErrVideoNotExist, err)
}

func TestFavoriteActionErr(t *testing.T) {
	err := FavoriteAction(12182603931062272, 10760595804524544, 1)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	err = FavoriteAction(12182603931062272, 10760595804524544, 1)
	assert.Equal(t, ErrAlreadyFavorite, err)
	err = FavoriteAction(12182603931062272, 10760595804524544, -1)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	err = FavoriteAction(12182603931062272, 10760595804524544, -1)
	assert.Equal(t, ErrNotFavorite, err)
}

func TestGetFavoriteList(t *testing.T) {
	_, err := GetFavoriteList(12182603931062272)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

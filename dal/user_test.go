package dal

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserByID(t *testing.T) {
	user, err := GetUserByID(context.Background(), 12182603931062272)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.NotNil(t, user)
}

func TestGetUserByIDs(t *testing.T) {
	users, err := GetUserByIDs(context.Background(), []int64{12182603931062272, 10760536648060928})
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.NotNil(t, users)
}

func TestGetUserByName(t *testing.T) {
	user := GetUserByName(context.Background(), "test01")
	assert.NotNil(t, user)
}

func TestCreateUser(t *testing.T) {
	err := CreateUser(context.Background(), "test02", "123456", 1111)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

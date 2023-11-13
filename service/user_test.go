package service

import (
	"douyin/config"
	"douyin/dao"
	"douyin/pkg/snowflake"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	snowflake.Init(&config.SnowflakeConfig{StartTime: "2023-10-10", MachineID: 1})
}
func TestUserInfo(t *testing.T) {
	resp, err := UserInfo(10760536648060928, 10760536648060928)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 0, int(resp.StatusCode))
	assert.Equal(t, int64(10760536648060928), resp.User.ID)
}

func TestRegister(t *testing.T) {
	resp, err := Register("test01", "123456")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 0, int(resp.StatusCode))
}

func TestLoginSuccess(t *testing.T) {
	resp, err := Login("test01", "123456")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 0, int(resp.StatusCode))
}

func TestLoginFail(t *testing.T) {
	_, err := Login("test32", "123456")
	assert.Equal(t, dao.ErrUserNotExist, err)
}

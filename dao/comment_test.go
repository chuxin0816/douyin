package dao

import (
	"douyin/config"
	"douyin/pkg/snowflake"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	conf := &config.DatabaseConfig{
		MysqlConfig: &config.MysqlConfig{
			User:     "root",
			Password: "123456",
			Host:     "127.0.0.1",
			Port:     3306,
			DBName:   "douyin",
		},
		RedisConfig: &config.RedisConfig{
			Host:     "127.0.0.1",
			Port:     6379,
			Password: "",
			DB:       1,
		},
	}
	if err := Init(conf); err != nil {
		panic(err)
	}
	if err := snowflake.Init(&config.SnowflakeConfig{StartTime: "2023-10-10", MachineID: 1}); err != nil {
		panic(err)
	}
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
	}
	assert.Equal(t, comment.ID, int64(12564221107638272))
}

func TestGetCommentList(t *testing.T) {
	comments, err := GetCommentList(10760595804524544)
	if err != nil {
		t.Log(err)
	}
	assert.Equal(t, 1, len(comments))
}

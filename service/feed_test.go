package service

import (
	"douyin/config"
	"douyin/dao"
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
	if err := dao.Init(conf); err != nil {
		panic(err)
	}
}

func TestFeed(t *testing.T) {
	resp, err := Feed(0, 10760536648060928)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 0, int(resp.StatusCode))
}

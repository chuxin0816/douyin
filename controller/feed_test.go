package controller

import (
	"douyin/config"
	"douyin/dao"
	"douyin/response"
	"encoding/json"
	"testing"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/ut"
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
	h := server.Default()
	h.GET("/feed", Feed)
	w := ut.PerformRequest(h.Engine, "GET", "/feed", nil)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode())
	result := &response.FeedResponse{}
	json.Unmarshal(resp.Body(), result)
	assert.Equal(t, 0, int(result.StatusCode))
}

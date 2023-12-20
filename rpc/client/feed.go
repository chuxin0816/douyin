package client

import (
	"douyin/config"
	"douyin/rpc/kitex_gen/feed"
	"douyin/rpc/kitex_gen/feed/feedservice"

	"github.com/cloudwego/kitex/client"
	consul "github.com/kitex-contrib/registry-consul"
)

var feedClient feedservice.Client

func initFeedClient(config *config.ConsulConfig) {
	// 服务发现
	r, err := consul.NewConsulResolver(config.Addr)
	if err != nil {
		panic(err)
	}

	feedClient, err = feedservice.NewClient(config.FeedServiceName, client.WithResolver(r))
	if err != nil {
		panic(err)
	}
}

func Feed(latestTime, userID int64) (*feed.FeedResponse, error) {

}

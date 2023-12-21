package snowflake

import (
	"douyin/config"
	"time"

	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

func Init() {
	st, err := time.Parse("2006-01-02", config.Conf.SnowflakeConfig.StartTime)
	if err != nil {
		panic(err)
	}
	snowflake.Epoch = st.UnixNano() / 1000000 //纳秒转毫秒
	node, err = snowflake.NewNode(config.Conf.SnowflakeConfig.MachineID)
	if err != nil {
		panic(err)
	}
}

func GenerateID() int64 {
	return node.Generate().Int64()
}

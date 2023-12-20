package snowflake

import (
	"douyin/config"
	"time"

	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

func Init(conf *config.SnowflakeConfig) {
	st, err := time.Parse("2006-01-02", conf.StartTime)
	if err != nil {
		panic(err)
	}
	snowflake.Epoch = st.UnixNano() / 1000000 //纳秒转毫秒
	node, err = snowflake.NewNode(conf.MachineID)
	if err != nil {
		panic(err)
	}
}

func GenerateID() int64 {
	return node.Generate().Int64()
}

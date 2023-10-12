package snowflake

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

var node *snowflake.Node

func Init(startTime string, machineID int64) error {
	st, err := time.Parse("2006-01-02", startTime)
	if err != nil {
		hlog.Error("snowflake.Init: 解析开始时间失败", err)
		return err
	}
	snowflake.Epoch = st.UnixNano() / 1000000 //纳秒转毫秒
	node, err = snowflake.NewNode(machineID)
	if err != nil {
		hlog.Error("snowflake.Init: 创建node失败", err)
		return err
	}
	return nil
}

func GenerateID() int64 {
	return node.Generate().Int64()
}

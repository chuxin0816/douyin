package logger

import (
	"os"
	"path"
	"time"

	"douyin/src/config"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/kitex/pkg/klog"
	hertzzap "github.com/hertz-contrib/logger/zap"
	kitexzap "github.com/kitex-contrib/obs-opentelemetry/logging/zap"

	"gopkg.in/natefinch/lumberjack.v2"
)

func Init() {
	// 可定制的输出目录。
	var logFilePath string = config.Conf.LogConfig.Path
	if err := os.MkdirAll(logFilePath, 0o777); err != nil {
		panic(err)
	}

	// 将文件名设置为日期
	logFileName := time.Now().Format("2006-01-02") + ".log"
	fileName := path.Join(logFilePath, logFileName)

	// 如果文件不存在，则创建一个新文件
	if _, err := os.Stat(fileName); err != nil {
		if _, err := os.Create(fileName); err != nil {
			panic(err)
		}
	}

	// 提供压缩和删除
	lumberjackLogger := &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    config.Conf.LogConfig.MaxSize,    // 一个文件最大可达 10M。
		MaxBackups: config.Conf.LogConfig.MaxBackups, // 最多同时保存 10 个文件。
		MaxAge:     config.Conf.LogConfig.MaxAge,     // 一个文件最多可以保存 30 天。
		Compress:   true,                             // 用 gzip 压缩。
	}

	hlog.SetLogger(hertzzap.NewLogger())
	hlog.SetLevel(hlog.LevelDebug)
	hlog.SetOutput(lumberjackLogger)
	klog.SetLogger(kitexzap.NewLogger())
	klog.SetLevel(klog.LevelDebug)
	klog.SetOutput(lumberjackLogger)
}

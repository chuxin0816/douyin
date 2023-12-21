package logger

import (
	"douyin/config"
	"os"
	"path"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	kitexlogrus "github.com/kitex-contrib/obs-opentelemetry/logging/logrus"

	"gopkg.in/natefinch/lumberjack.v2"
)

func Init(conf *config.LogConfig) {
	// 可定制的输出目录。
	var logFilePath string = conf.Path
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
		MaxSize:    conf.MaxSize,    // 一个文件最大可达 10M。
		MaxBackups: conf.MaxBackups, // 最多同时保存 10 个文件。
		MaxAge:     conf.MaxAge,     // 一个文件最多可以保存 30 天。
		Compress:   true,            // 用 gzip 压缩。
	}

	klog.SetLogger(kitexlogrus.NewLogger())
	klog.SetLevel(klog.LevelDebug)
	klog.SetOutput(lumberjackLogger)
}

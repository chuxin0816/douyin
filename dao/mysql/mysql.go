package mysql

import (
	"douyin/models"
	"fmt"

	"github.com/chuxin0816/Scaffold/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func Init(conf *config.MysqlConfig) (err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.User,
		conf.Password,
		conf.Host,
		conf.Port,
		conf.DBName,
	)

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent)}) // 禁用gorm日志
	if err != nil {
		hlog.Error("mysql.Init: 连接数据库失败")
		return err
	}

	err = db.AutoMigrate(&models.User{}, &models.Video{})
	if err != nil {
		hlog.Error("mysql.Init: 数据库迁移失败")
		return err
	}
	return nil
}

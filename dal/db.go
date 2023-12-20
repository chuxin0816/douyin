package dal

import (
	"douyin/config"
	"douyin/dal/query"
	"fmt"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB  *gorm.DB
	RDB *redis.Client
)

func Init(config *config.DatabaseConfig) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.MysqlConfig.User, config.MysqlConfig.Password, config.MysqlConfig.Host, config.MysqlConfig.Port, config.MysqlConfig.DBName)

	DB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.RedisConfig.Host, config.RedisConfig.Port),
		Password: config.RedisConfig.Password,
		DB:       config.RedisConfig.DB,
	})
	query.SetDefault(DB)
}

func Close() {
	RDB.Close()
}

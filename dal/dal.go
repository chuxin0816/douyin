package dal

import (
	"douyin/config"
	"douyin/dal/query"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB  *gorm.DB
	RDB *redis.Client
)

const (
	expireTime = time.Hour * 72
	timeout    = time.Second * 5
	delayTime  = 100 * time.Millisecond
	randFactor = 30
	tickerTime = time.Second * 10
)

var (
	ErrUserExist       = errors.New("用户已存在")
	ErrUserNotExist    = errors.New("用户不存在")
	ErrPassword        = errors.New("密码错误")
	ErrAlreadyFollow   = errors.New("已经关注过了")
	ErrNotFollow       = errors.New("还没有关注过")
	ErrAlreadyFavorite = errors.New("已经点赞过了")
	ErrNotFavorite     = errors.New("还没有点赞过")
	ErrCommentNotExist = errors.New("comment not exist")
	ErrVideoNotExist   = errors.New("video not exist")
)

func Init() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=loc=Asia%%2FShanghai",
		config.Conf.DatabaseConfig.MysqlConfig.User, config.Conf.DatabaseConfig.MysqlConfig.Password, config.Conf.DatabaseConfig.MysqlConfig.Host, config.Conf.DatabaseConfig.MysqlConfig.Port, config.Conf.DatabaseConfig.MysqlConfig.DBName)

	DB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	RDB = redis.NewClient(&redis.Options{
		Addr:     config.Conf.DatabaseConfig.RedisConfig.Addr,
		Password: config.Conf.DatabaseConfig.RedisConfig.Password,
		DB:       config.Conf.DatabaseConfig.RedisConfig.DB,
	})
	query.SetDefault(DB)
}

func Close() {
	RDB.Close()
}

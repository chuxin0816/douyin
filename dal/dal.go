package dal

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"douyin/config"
	"douyin/dal/query"

	"github.com/bits-and-blooms/bloom/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

const (
	ExpireTime = time.Hour * 72
	delayTime  = 100 * time.Millisecond
	randFactor = 30
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

var (
	db                *gorm.DB
	RDB               *redis.Client
	collectionMessage *mongo.Collection
	g                 *singleflight.Group
	bloomFilter       *bloom.BloomFilter
	CacheUserID       sync.Map
	CacheVideoID      sync.Map
)

// nil值，用于占位，于Init函数中初始化
var (
	qComment  = query.Comment
	qFavorite = query.Favorite
	qRelation = query.Relation
	qUser     = query.User
	qVideo    = query.Video
)

func Init() {
	// 初始化MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Conf.DatabaseConfig.MysqlConfig.User,
		config.Conf.DatabaseConfig.MysqlConfig.Password,
		config.Conf.DatabaseConfig.MysqlConfig.Host,
		config.Conf.DatabaseConfig.MysqlConfig.Port,
		config.Conf.DatabaseConfig.MysqlConfig.DBName,
	)

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	if err := db.Use(tracing.NewPlugin(tracing.WithoutMetrics())); err != nil {
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxIdleConns(100)
	sqlDB.SetMaxOpenConns(200)
	sqlDB.SetConnMaxLifetime(24 * time.Hour)
	sqlDB.SetConnMaxIdleTime(time.Hour)

	query.SetDefault(db)
	qComment = query.Comment
	qFavorite = query.Favorite
	qRelation = query.Relation
	qUser = query.User
	qVideo = query.Video

	// 初始化Redis
	RDB = redis.NewClient(&redis.Options{
		Addr:     config.Conf.DatabaseConfig.RedisConfig.Addr,
		Password: config.Conf.DatabaseConfig.RedisConfig.Password,
		DB:       config.Conf.DatabaseConfig.RedisConfig.DB,
	})
	if err := redisotel.InstrumentTracing(RDB); err != nil {
		panic(err)
	}

	if err := RDB.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}

	// 初始化MongoDB
	uri := fmt.Sprintf("mongodb://%s:%d", config.Conf.MongoConfig.Host, config.Conf.MongoConfig.Port)
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		panic(err)
	}

	collectionMessage = client.Database(config.Conf.MongoConfig.DBName).Collection("message")

	// 初始化singleflight
	g = &singleflight.Group{}

	// 初始化布隆过滤器
	bloomFilter = bloom.NewWithEstimates(100000, 0.001)
	if err := loadDataToBloom(); err != nil {
		panic(err)
	}
}

func Close() {
	RDB.Close()
}

func RemoveFavoriteCache(ctx context.Context, userID, videoID string) error {
	key := GetRedisKey(KeyUserFavoritePF + userID)
	return RDB.SRem(ctx, key, videoID).Err()
}

func RemoveRelationCache(ctx context.Context, userID, toUserID string) error {
	key := GetRedisKey(KeyUserFollowerPF + toUserID)
	return RDB.SRem(ctx, key, userID).Err()
}

// GetRandomTime 获取0-30min随机时间
func GetRandomTime() time.Duration {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return time.Duration(r.Intn(randFactor)) * time.Minute
}

func loadDataToBloom() error {
	// 填入用户ID和name
	PageSize := 30
	PageCnt := 0
	cnt, err := qUser.WithContext(context.Background()).Count()
	if err != nil {
		return err
	}
	count := int(cnt)
	if count%PageSize == 0 {
		PageCnt = count / PageSize
	} else {
		PageCnt = count/PageSize + 1
	}

	for page := 0; page < PageCnt; page++ {
		users, err := qUser.WithContext(context.Background()).
			Offset(PageSize*page).Limit(PageSize).
			Select(qUser.ID, qUser.Name).Find()
		if err != nil {
			return err
		}

		for _, user := range users {
			bloomFilter.Add([]byte(strconv.FormatInt(user.ID, 10)))
			bloomFilter.Add([]byte(user.Name))
		}
	}

	// 填入视频ID
	cnt, err = qVideo.WithContext(context.Background()).Count()
	if err != nil {
		return err
	}
	count = int(cnt)
	if count%PageSize == 0 {
		PageCnt = count / PageSize
	} else {
		PageCnt = count/PageSize + 1
	}

	for page := 0; page < PageCnt; page++ {
		videos, err := qVideo.WithContext(context.Background()).
			Offset(PageSize * page).Limit(PageSize).
			Select(qVideo.ID).Find()
		if err != nil {
			return err
		}

		for _, video := range videos {
			bloomFilter.Add([]byte(strconv.FormatInt(video.ID, 10)))
		}
	}

	return nil
}

func AddToBloom(ID string) {
	bloomFilter.Add([]byte(ID))
}

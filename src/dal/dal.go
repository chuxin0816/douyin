package dal

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"douyin/src/config"
	"douyin/src/dal/query"

	"github.com/bits-and-blooms/bloom/v3"
	nebula "github.com/vesoft-inc/nebula-go/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	gormTracing "gorm.io/plugin/opentelemetry/tracing"
)

const (
	ExpireTime = time.Hour * 24
	DelayTime  = 150 * time.Millisecond
	randFactor = 30
)

var (
	ErrUserExist       = errors.New("用户已存在")
	ErrUserNotExist    = errors.New("用户不存在")
	ErrPassword        = errors.New("密码错误")
	ErrAlreadyFollow   = errors.New("已经关注过了")
	ErrNotFollow       = errors.New("还没有关注过")
	ErrFollowLimit     = errors.New("关注数超过限制")
	ErrAlreadyFavorite = errors.New("已经点赞过了")
	ErrNotFavorite     = errors.New("还没有点赞过")
	ErrCommentNotExist = errors.New("comment not exist")
	ErrVideoNotExist   = errors.New("video not exist")
)

var (
	db                *gorm.DB
	RDB               *redis.ClusterClient
	collectionMessage *mongo.Collection
	sessionPool       *nebula.SessionPool
	G                 = &singleflight.Group{}
	bloomFilter       *bloom.BloomFilter
)

var (
	q          = new(query.Query)
	qComment   = q.Comment
	qFavorite  = q.Favorite
	qUser      = q.User
	qUserLogin = q.UserLogin
	qVideo     = q.Video
)

func Init() {
	// 初始化MySQL
	InitMySQL()

	// 初始化Redis
	InitRedis()

	// 初始化MongoDB
	InitMongo()

	// 初始化Nebula
	InitNebula()

	// 初始化布隆过滤器
	bloomFilter = bloom.NewWithEstimates(100000, 0.001)
	if err := loadDataToBloom(); err != nil {
		panic(err)
	}
}

func InitMySQL() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Conf.DatabaseConfig.MySQLMaster.User,
		config.Conf.DatabaseConfig.MySQLMaster.Password,
		config.Conf.DatabaseConfig.MySQLMaster.Host,
		config.Conf.DatabaseConfig.MySQLMaster.Port,
		config.Conf.DatabaseConfig.MySQLMaster.DBName,
	)

	// 连接主库
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// 连接从库
	replicas := make([]gorm.Dialector, len(config.Conf.DatabaseConfig.MySQLSlaves))
	for i, slave := range config.Conf.DatabaseConfig.MySQLSlaves {
		replicas[i] = mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			slave.User,
			slave.Password,
			slave.Host,
			slave.Port,
			slave.DBName,
		))
	}
	err = db.Use(
		dbresolver.Register(dbresolver.Config{
			Replicas: replicas,
			Policy:   dbresolver.RandomPolicy{},
		}).
			// 设置连接池
			SetConnMaxIdleTime(time.Minute * 30).
			SetConnMaxLifetime(time.Hour).
			SetMaxIdleConns(100).
			SetMaxOpenConns(500),
	)
	if err != nil {
		panic(err)
	}

	// 链路追踪插件
	if err := db.Use(gormTracing.NewPlugin(gormTracing.WithoutMetrics())); err != nil {
		panic(err)
	}

	q = query.Use(db)
	qComment = q.Comment
	qFavorite = q.Favorite
	qUser = q.User
	qUserLogin = q.UserLogin
	qVideo = q.Video
}

func InitRedis() {
	RDB = redis.NewFailoverClusterClient(&redis.FailoverOptions{
		MasterName:    config.Conf.DatabaseConfig.Redis.MasterName,
		SentinelAddrs: config.Conf.DatabaseConfig.Redis.SentinelAddrs,
		Password:      config.Conf.DatabaseConfig.Redis.Password,
		DB:            config.Conf.DatabaseConfig.Redis.DB,
	})
	if err := RDB.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}

	if err := redisotel.InstrumentTracing(RDB); err != nil {
		panic(err)
	}
}

func InitMongo() {
	uri := fmt.Sprintf("mongodb://%s:%d", config.Conf.Mongo.Host, config.Conf.Mongo.Port)
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		panic(err)
	}

	collectionMessage = client.Database(config.Conf.Mongo.DBName).Collection("message")
}

func InitNebula() {
	hostAddr := nebula.HostAddress{Host: config.Conf.DatabaseConfig.Nebula.Host, Port: config.Conf.DatabaseConfig.Nebula.Port}
	cfg, err := nebula.NewSessionPoolConf(
		config.Conf.DatabaseConfig.Nebula.User,
		config.Conf.DatabaseConfig.Nebula.Password,
		[]nebula.HostAddress{hostAddr},
		config.Conf.DatabaseConfig.Nebula.Space,
	)
	if err != nil {
		panic(err)
	}

	sessionPool, err = nebula.NewSessionPool(*cfg, nebula.DefaultLogger{})
	if err != nil {
		panic(err)
	}
}

func Close() {
	RDB.Close()
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

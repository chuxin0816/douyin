package dal

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"douyin/src/config"
	"douyin/src/dal/model"
	"douyin/src/dal/query"
	"douyin/src/pkg/tracing"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/cloudwego/kitex/pkg/klog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/codes"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormTracing "gorm.io/plugin/opentelemetry/tracing"
)

const (
	ExpireTime        = time.Hour * 72
	delayTime         = 100 * time.Millisecond
	randFactor        = 30
	aggregateInterval = time.Second * 2
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
	RDB               *redis.Client
	collectionMessage *mongo.Collection
	g                 = &singleflight.Group{}
	bloomFilter       *bloom.BloomFilter
	CacheUserID       = make(map[int64]struct{})
	CacheVideoID      = make(map[int64]struct{})
	Mu                sync.Mutex
)

var (
	q         = new(query.Query)
	qComment  = q.Comment
	qFavorite = q.Favorite
	qRelation = q.Relation
	qUser     = q.User
	qVideo    = q.Video
)

func Init() {
	// 初始化MySQL
	InitMySQL()

	// 初始化Redis
	InitRedis()

	// 初始化MongoDB
	InitMongo()

	// 初始化布隆过滤器
	bloomFilter = bloom.NewWithEstimates(100000, 0.001)
	if err := loadDataToBloom(); err != nil {
		panic(err)
	}

	go syncRedisToMySQL(context.Background())
}

func InitMySQL() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Conf.DatabaseConfig.MySQLConfig.User,
		config.Conf.DatabaseConfig.MySQLConfig.Password,
		config.Conf.DatabaseConfig.MySQLConfig.Host,
		config.Conf.DatabaseConfig.MySQLConfig.Port,
		config.Conf.DatabaseConfig.MySQLConfig.DBName,
	)

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	if err := db.Use(gormTracing.NewPlugin(gormTracing.WithoutMetrics())); err != nil {
		panic(err)
	}

	q = query.Use(db)
	qComment = q.Comment
	qFavorite = q.Favorite
	qRelation = q.Relation
	qUser = q.User
	qVideo = q.Video

	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxIdleConns(100)
	sqlDB.SetMaxOpenConns(500)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(time.Minute * 30)
}

func InitRedis() {
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
}

func InitMongo() {
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
}

func Close() {
	RDB.Close()
}

func syncRedisToMySQL(ctx context.Context) {
	ticker := time.NewTicker(aggregateInterval)
	defer ticker.Stop()
	for {
		<-ticker.C
		go syncUser(ctx)
		go syncVideo(ctx)
	}
}

func syncUser(ctx context.Context) {
	if len(CacheUserID) == 0 {
		return
	}

	ctx, span := tracing.Tracer.Start(ctx, "syncUser")
	defer span.End()

	// 备份缓存中的用户ID并清空
	Mu.Lock()
	backup := CacheUserID
	CacheUserID = make(map[int64]struct{})
	Mu.Unlock()

	// 同步redis的用户缓存到Mysql
	pipe := RDB.Pipeline()

	for userID := range backup {
		userIDStr := strconv.FormatInt(userID, 10)
		pipe.Get(ctx, GetRedisKey(KeyUserTotalFavoritedPF+userIDStr))
		pipe.Get(ctx, GetRedisKey(KeyUserFavoriteCountPF+userIDStr))
		pipe.Get(ctx, GetRedisKey(KeyUserFollowCountPF+userIDStr))
		pipe.Get(ctx, GetRedisKey(KeyUserFollowerCountPF+userIDStr))
		pipe.Get(ctx, GetRedisKey(KeyUserWorkCountPF+userIDStr))

		cmds, err := pipe.Exec(ctx)
		if err != nil && err != redis.Nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to exec pipeline")
			klog.Error("同步redis用户缓存到mysql失败,err: ", err)
			return
		}

		totalFavorited, _ := strconv.ParseInt(cmds[0].(*redis.StringCmd).Val(), 10, 64)
		favoriteCount, _ := strconv.ParseInt(cmds[1].(*redis.StringCmd).Val(), 10, 64)
		followCount, _ := strconv.ParseInt(cmds[2].(*redis.StringCmd).Val(), 10, 64)
		followerCount, _ := strconv.ParseInt(cmds[3].(*redis.StringCmd).Val(), 10, 64)
		workCount, _ := strconv.ParseInt(cmds[4].(*redis.StringCmd).Val(), 10, 64)
		mUser := &model.User{
			ID:             userID,
			TotalFavorited: totalFavorited,
			FavoriteCount:  favoriteCount,
			FollowCount:    followCount,
			FollowerCount:  followerCount,
			WorkCount:      workCount,
		}
		if err := UpdateUser(ctx, mUser); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "同步redis用户缓存到mysql失败")
			klog.Error("同步redis用户缓存到mysql失败,err: ", err)
			continue
		}
	}
}

func syncVideo(ctx context.Context) {
	if len(CacheVideoID) == 0 {
		return
	}

	ctx, span := tracing.Tracer.Start(ctx, "syncVideo")
	defer span.End()

	// 备份缓存中的视频ID并清空
	Mu.Lock()
	backup := CacheVideoID
	CacheVideoID = make(map[int64]struct{})
	Mu.Unlock()

	// 同步redis中的视频缓存到Mysql
	pipe := RDB.Pipeline()

	for videoID := range backup {
		videoIDStr := strconv.FormatInt(videoID, 10)
		pipe.Get(ctx, GetRedisKey(KeyVideoFavoriteCountPF+videoIDStr))
		pipe.Get(ctx, GetRedisKey(KeyVideoCommentCountPF+videoIDStr))

		cmds, err := pipe.Exec(ctx)
		if err != nil && err != redis.Nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "同步redis视频缓存到mysql失败")
			klog.Errorf("同步redis视频缓存到mysql失败,err: ", err)
			continue
		}

		videoFavoriteCount, _ := strconv.ParseInt(cmds[0].(*redis.StringCmd).Val(), 10, 64)
		videoCommentCount, _ := strconv.ParseInt(cmds[1].(*redis.StringCmd).Val(), 10, 64)
		mVideo := &model.Video{
			ID:            videoID,
			FavoriteCount: videoFavoriteCount,
			CommentCount:  videoCommentCount,
		}
		if err := UpdateVideo(ctx, mVideo); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "同步redis视频缓存到mysql失败")
			klog.Errorf("同步redis视频缓存到mysql失败,err: ", err)
			continue
		}
	}
}

func RemoveFavoriteCache(ctx context.Context, userID, videoID string) error {
	key := GetRedisKey(KeyUserFavoritePF + userID)
	return RDB.SRem(ctx, key, videoID).Err()
}

func RemoveRelationCache(ctx context.Context, userID, toUserID string) error {
	keyFollowPF := GetRedisKey(KeyUserFollowPF + userID)
	keyFollower := GetRedisKey(KeyUserFollowerPF + toUserID)

	pipe := RDB.Pipeline()
	pipe.SRem(ctx, keyFollowPF, toUserID)
	pipe.SRem(ctx, keyFollower, userID)
	_, err := pipe.Exec(ctx)

	return err
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

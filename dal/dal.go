package dal

import (
	"context"
	"douyin/config"
	"douyin/dal/query"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/cloudwego/kitex/pkg/klog"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	expireTime        = time.Hour * 72
	timeout           = time.Second * 5
	delayTime         = 100 * time.Second
	randFactor        = 30
	aggregateInterval = time.Second * 10
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
	db           *gorm.DB
	rdb          *redis.Client
	g            *singleflight.Group
	bloomFilter  *bloom.BloomFilter
	cacheUserID  sync.Map
	cacheVideoID sync.Map
)

var (
	qComment  = query.Comment
	qFavorite = query.Favorite
	qMessage  = query.Message
	qRelation = query.Relation
	qUser     = query.User
	qVideo    = query.Video
)

func Init() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FShanghai",
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
	query.SetDefault(db)

	rdb = redis.NewClient(&redis.Options{
		Addr:     config.Conf.DatabaseConfig.RedisConfig.Addr,
		Password: config.Conf.DatabaseConfig.RedisConfig.Password,
		DB:       config.Conf.DatabaseConfig.RedisConfig.DB,
	})
	err = rdb.Ping(context.Background()).Err()
	if err != nil {
		panic(err)
	}

	// 初始化singleflight
	g = &singleflight.Group{}

	// 初始化布隆过滤器
	bloomFilter = bloom.NewWithEstimates(100000, 0.001)
	if err := loadDataToBloom(); err != nil {
		panic(err)
	}

	// 开启定时同步任务
	go syncRedisToMySQL()
}

func Close() {
	rdb.Close()
}

func syncRedisToMySQL() {
	ticker := time.NewTicker(aggregateInterval)
	defer ticker.Stop()
	for {
		<-ticker.C

		// 备份缓存中的用户ID和视频ID并清空
		backupUserID := make(map[int64]struct{})
		backupVideoID := make(map[int64]struct{})

		cacheUserID.Range(func(key, value any) bool {
			backupUserID[key.(int64)] = struct{}{}
			cacheUserID.Delete(key)
			return true
		})
		cacheVideoID.Range(func(key, value any) bool {
			backupVideoID[key.(int64)] = struct{}{}
			cacheVideoID.Delete(key)
			return true
		})

		// 同步redis的用户缓存到Mysql
		for userID := range backupUserID {
			userIDStr := strconv.FormatInt(userID, 10)
			totalFavorited, _ := strconv.ParseInt(rdb.Get(context.Background(), getRedisKey(KeyUserTotalFavoritedPF+userIDStr)).Val(), 10, 64)
			favoriteCount, _ := strconv.ParseInt(rdb.Get(context.Background(), getRedisKey(KeyUserFavoriteCountPF+userIDStr)).Val(), 10, 64)
			followCount, _ := strconv.ParseInt(rdb.Get(context.Background(), getRedisKey(KeyUserFollowCountPF+userIDStr)).Val(), 10, 64)
			followerCount, _ := strconv.ParseInt(rdb.Get(context.Background(), getRedisKey(KeyUserFollowerCountPF+userIDStr)).Val(), 10, 64)
			workCount, _ := strconv.ParseInt(rdb.Get(context.Background(), getRedisKey(KeyUserWorkCountPF+userIDStr)).Val(), 10, 64)
			_, err := qUser.WithContext(context.Background()).Where(qUser.ID.Eq(userID)).Updates(map[string]interface{}{
				"total_favorited": totalFavorited,
				"favorite_count":  favoriteCount,
				"follow_count":    followCount,
				"follower_count":  followerCount,
				"work_count":      workCount,
			})
			if err != nil {
				klog.Error("同步redis用户缓存到mysql失败")
				continue
			}
		}

		// 同步redis中的视频缓存到Mysql
		var videoFavoriteCount, videoCommentCount int64
		for videoID := range backupVideoID {
			videoIDStr := strconv.FormatInt(videoID, 10)
			videoFavoriteCount, _ = strconv.ParseInt(rdb.Get(context.Background(), getRedisKey(KeyVideoFavoriteCountPF+videoIDStr)).Val(), 10, 64)
			videoCommentCount, _ = strconv.ParseInt(rdb.Get(context.Background(), getRedisKey(KeyVideoCommentCountPF+videoIDStr)).Val(), 10, 64)
			_, err := qVideo.WithContext(context.Background()).Where(qVideo.ID.Eq(videoID)).Updates(map[string]interface{}{
				"favorite_count": videoFavoriteCount,
				"comment_count":  videoCommentCount,
			})
			if err != nil {
				klog.Error("同步redis视频缓存到mysql失败")
				continue
			}
		}
	}
}

func getRandomTime() time.Duration {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return time.Duration(r.Intn(randFactor)) * time.Minute
}

func loadDataToBloom() error {
	// 填入用户ID和name
	PageSize := 30
	PageCnt := 0
	cnt, err := qUser.WithContext(context.Background()).Count()
	if err != nil {
		klog.Error("查询用户数量失败")
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
			klog.Error("查询用户名和id失败")
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
		klog.Error("查询视频数量失败")
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
			klog.Error("查询视频id失败")
			return err
		}

		for _, video := range videos {
			bloomFilter.Add([]byte(strconv.FormatInt(video.ID, 10)))
		}
	}

	return nil
}

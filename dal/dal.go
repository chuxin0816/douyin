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
	ExpireTime        = time.Hour * 72
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
	RDB          *redis.Client
	g            *singleflight.Group
	bloomFilter  *bloom.BloomFilter
	CacheUserID  sync.Map
	CacheVideoID sync.Map
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

	RDB = redis.NewClient(&redis.Options{
		Addr:     config.Conf.DatabaseConfig.RedisConfig.Addr,
		Password: config.Conf.DatabaseConfig.RedisConfig.Password,
		DB:       config.Conf.DatabaseConfig.RedisConfig.DB,
	})
	err = RDB.Ping(context.Background()).Err()
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
	RDB.Close()
}

func RemoveFavoriteCache(ctx context.Context, userID, videoID string) {
	key := GetRedisKey(KeyUserFavoritePF + userID)
	if err := RDB.SRem(ctx, key, videoID).Err(); err != nil {
		klog.Error("删除redis缓存失败, err: ", err)
	}
}

func RemoveRelationCache(ctx context.Context, userID, toUserID string) {
	key := GetRedisKey(KeyUserFollowerPF + toUserID)
	if err := RDB.SRem(ctx, key, userID).Err(); err != nil {
		klog.Error("删除redis缓存失败, err: ", err)
	}
}

func syncRedisToMySQL() {
	ticker := time.NewTicker(aggregateInterval)
	defer ticker.Stop()
	for {
		<-ticker.C
		go syncUser()
		go syncVideo()
	}
}

func syncUser() {
	// 备份缓存中的用户ID并清空
	backupUserID := make([]int64, 0, 100000)

	CacheUserID.Range(func(key, value any) bool {
		backupUserID = append(backupUserID, key.(int64))
		CacheUserID.Delete(key)
		return true
	})

	// 同步redis的用户缓存到Mysql
	pipe := RDB.Pipeline()

	for _, userID := range backupUserID {
		userIDStr := strconv.FormatInt(userID, 10)
		pipe.Get(context.Background(), GetRedisKey(KeyUserTotalFavoritedPF+userIDStr))
		pipe.Get(context.Background(), GetRedisKey(KeyUserFavoriteCountPF+userIDStr))
		pipe.Get(context.Background(), GetRedisKey(KeyUserFollowCountPF+userIDStr))
		pipe.Get(context.Background(), GetRedisKey(KeyUserFollowerCountPF+userIDStr))
		pipe.Get(context.Background(), GetRedisKey(KeyUserWorkCountPF+userIDStr))
	}

	cmds, err := pipe.Exec(context.Background())
	if err != nil && err != redis.Nil {
		klog.Error("同步redis用户缓存到mysql失败,err: ", err)
		return
	}
	for i, userID := range backupUserID {
		totalFavorited, _ := strconv.ParseInt(cmds[i*5].(*redis.StringCmd).Val(), 10, 64)
		favoriteCount, _ := strconv.ParseInt(cmds[i*5+1].(*redis.StringCmd).Val(), 10, 64)
		followCount, _ := strconv.ParseInt(cmds[i*5+2].(*redis.StringCmd).Val(), 10, 64)
		followerCount, _ := strconv.ParseInt(cmds[i*5+3].(*redis.StringCmd).Val(), 10, 64)
		workCount, _ := strconv.ParseInt(cmds[i*5+4].(*redis.StringCmd).Val(), 10, 64)
		_, err = qUser.WithContext(context.Background()).Where(qUser.ID.Eq(userID)).Updates(map[string]interface{}{
			"total_favorited": totalFavorited,
			"favorite_count":  favoriteCount,
			"follow_count":    followCount,
			"follower_count":  followerCount,
			"work_count":      workCount,
		})
		if err != nil {
			klog.Error("同步redis用户缓存到mysql失败")
			return
		}
	}
}
func syncVideo() {
	// 备份缓存中的视频ID并清空
	backupVideoID := make([]int64, 0, 100000)

	CacheVideoID.Range(func(key, value any) bool {
		backupVideoID = append(backupVideoID, key.(int64))
		CacheVideoID.Delete(key)
		return true
	})

	// 同步redis中的视频缓存到Mysql
	pipe := RDB.Pipeline()
	for i, videoID := range backupVideoID {
		videoIDStr := strconv.FormatInt(videoID, 10)
		pipe.Get(context.Background(), GetRedisKey(KeyVideoFavoriteCountPF+videoIDStr))
		pipe.Get(context.Background(), GetRedisKey(KeyVideoCommentCountPF+videoIDStr))
		cmds, err := pipe.Exec(context.Background())
		if err != nil {
			if err == redis.Nil {
				klog.Warnf("redis中不存在视频ID为%d的缓存", videoID)
				continue
			}
			klog.Errorf("同步redis视频缓存到mysql失败,err: ", err)
			continue
		}
		videoFavoriteCount, _ := strconv.ParseInt(cmds[i*2].(*redis.StringCmd).Val(), 10, 64)
		videoCommentCount, _ := strconv.ParseInt(cmds[i*2+1].(*redis.StringCmd).Val(), 10, 64)
		_, err = qVideo.WithContext(context.Background()).Where(qVideo.ID.Eq(videoID)).Updates(map[string]interface{}{
			"favorite_count": videoFavoriteCount,
			"comment_count":  videoCommentCount,
		})
		if err != nil {
			klog.Error("同步redis视频缓存到mysql失败")
			continue
		}
	}
}

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

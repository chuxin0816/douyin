package dao

import (
	"context"
	"douyin/config"
	"douyin/models"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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

var (
	db            *gorm.DB
	rdb           *redis.Client
	g             *singleflight.Group
	bloomFilter   *bloom.BloomFilter
	cacheUserID   []int64
	cacheVideoIDs []int64
	lock          sync.Mutex
)

func Init(conf *config.DatabaseConfig) (err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.User,
		conf.MysqlConfig.Password,
		conf.MysqlConfig.Host,
		conf.MysqlConfig.Port,
		conf.MysqlConfig.DBName,
	)

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		hlog.Error("mysql.Init: 连接数据库失败")
		return err
	}

	err = db.AutoMigrate(&models.User{}, &models.Video{}, &models.Favorite{}, &models.Comment{}, &models.Relation{}, &models.Message{})
	if err != nil {
		hlog.Error("mysql.Init: 数据库迁移失败")
		return err
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", conf.RedisConfig.Host, conf.RedisConfig.Port),
		Password: conf.RedisConfig.Password,
		DB:       conf.RedisConfig.DB,
	})
	err = rdb.Ping(context.Background()).Err()
	if err != nil {
		hlog.Error("redis.Init: 连接redis失败")
	}

	// 初始化singleflight
	g = &singleflight.Group{}

	// 初始化布隆过滤器
	bloomFilter = bloom.NewWithEstimates(100000, 0.001)
	if err := loadDataToBloom(); err != nil {
		hlog.Error("dao.Init: 加载数据到布隆过滤器失败")
		return err
	}

	// 开启定时同步任务
	go syncRedisToMySQL()
	return
}

func Close() {
	rdb.Close()
}

func syncRedisToMySQL() {
	ticker := time.NewTicker(tickerTime)
	defer ticker.Stop()
	for {
		<-ticker.C

		// 备份缓存中的用户ID和视频ID并清空
		lock.Lock()
		backupUserIDs := cacheUserID
		backupVideoIDs := cacheVideoIDs
		cacheUserID = cacheUserID[:0]
		cacheVideoIDs = cacheVideoIDs[:0]
		lock.Unlock()

		// 同步redis的用户缓存到Mysql
		for _, userID := range backupUserIDs {
			userIDStr := strconv.FormatInt(userID, 10)
			totalFavorited, _ := strconv.ParseInt(rdb.Get(context.Background(), getRedisKey(KeyUserTotalFavoritedPF+userIDStr)).Val(), 10, 64)
			favoriteCount, _ := strconv.ParseInt(rdb.Get(context.Background(), getRedisKey(KeyUserFavoriteCountPF+userIDStr)).Val(), 10, 64)
			followCount, _ := strconv.ParseInt(rdb.Get(context.Background(), getRedisKey(KeyUserFollowCountPF+userIDStr)).Val(), 10, 64)
			followerCount, _ := strconv.ParseInt(rdb.Get(context.Background(), getRedisKey(KeyUserFollowerCountPF+userIDStr)).Val(), 10, 64)
			workCount, _ := strconv.ParseInt(rdb.Get(context.Background(), getRedisKey(KeyUserWorkCountPF+userIDStr)).Val(), 10, 64)
			err := db.Table("users").Where("id = ?", userID).Updates(map[string]interface{}{
				"total_favorited": totalFavorited,
				"favorite_count":  favoriteCount,
				"follow_count":    followCount,
				"follower_count":  followerCount,
				"work_count":      workCount}).Error
			if err != nil {
				hlog.Error("dao.syncRedisToMySQL: 同步redis用户缓存到mysql失败")
				continue
			}
		}

		// 同步redis中的视频缓存到Mysql
		var videoFavoriteCount, videoCommentCount int64
		for _, videoID := range backupVideoIDs {
			videoIDStr := strconv.FormatInt(videoID, 10)
			videoFavoriteCount, _ = strconv.ParseInt(rdb.Get(context.Background(), getRedisKey(KeyVideoFavoriteCountPF+videoIDStr)).Val(), 10, 64)
			videoCommentCount, _ = strconv.ParseInt(rdb.Get(context.Background(), getRedisKey(KeyVideoCommentCountPF+videoIDStr)).Val(), 10, 64)
			err := db.Table("videos").Where("id = ?", videoID).Updates(map[string]interface{}{
				"favorite_count": videoFavoriteCount,
				"comment_count":  videoCommentCount}).Error
			if err != nil {
				hlog.Error("dao.syncRedisToMySQL: 同步redis视频缓存到mysql失败")
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
	var users []*models.User
	PageSize := 50
	PageCnt := 0
	var Cnt int64
	if err := db.Model(&models.User{}).Count(&Cnt).Error; err != nil {
		hlog.Error("dao.loadDataToBloom: 查询用户数量失败")
		return err
	}
	count := int(Cnt)
	if count%PageSize == 0 {
		PageCnt = count / PageSize
	} else {
		PageCnt = count/PageSize + 1
	}
	
	for page := 0; page < PageCnt; page++ {
		if err := db.Select("id", "name").Offset(PageSize * page).Limit(PageSize).Find(&users).Error; err != nil {
			hlog.Error("dao.loadDataToBloom: 查询用户名和id失败")
			return err
		}

		for _, user := range users {
			bloomFilter.Add([]byte(strconv.FormatInt(user.ID, 10)))
			bloomFilter.Add([]byte(user.Name))
		}
	}

	// 填入视频ID
	var videoIDs []int64
	if err := db.Model(&models.Video{}).Count(&Cnt).Error; err != nil {
		hlog.Error("dao.loadDataToBloom: 查询视频数量失败")
		return err
	}
	count = int(Cnt)
	if count%PageSize == 0 {
		PageCnt = count / PageSize
	} else {
		PageCnt = count/PageSize + 1
	}

	for page := 0; page < PageCnt; page++ {
		if err := db.Model(&models.Video{}).Select("id").Offset(PageSize * page).Limit(PageSize).Find(&videoIDs).Error; err != nil {
			hlog.Error("dao.loadDataToBloom: 查询视频id失败")
			return err
		}

		for _, videoID := range videoIDs {
			bloomFilter.Add([]byte(strconv.FormatInt(videoID, 10)))
		}
	}

	return nil
}

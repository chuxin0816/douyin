package dao

import (
	"context"
	"douyin/config"
	"douyin/models"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db          *gorm.DB
	rdb         *redis.Client
	bloomFilter *bloom.BloomFilter
	userIDs     []int64
	videoIDs    []int64
	rwLock      sync.RWMutex
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

	bloomFilter = bloom.NewWithEstimates(100000, 0.001)
	// 初始化布隆过滤器
	var usernames []string
	if err = db.Table("users").Select("name").Find(&usernames).Error; err != nil {
		hlog.Error("dao.Init: 查询用户名失败")
		return err
	}
	for _, username := range usernames {
		bloomFilter.Add([]byte(username))
	}
	// 开启定时同步任务
	go syncRedisToMySQL()
	return
}

func Close() {
	rdb.Close()
}

func syncRedisToMySQL() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		<-ticker.C

		// 备份缓存中的用户ID和视频ID并清空
		rwLock.Lock()
		backupUserIDs := make([]int64, len(userIDs))
		backupVideoIDs := make([]int64, len(videoIDs))
		copy(backupUserIDs, userIDs)
		copy(backupVideoIDs, videoIDs)
		userIDs = userIDs[:0]
		videoIDs = videoIDs[:0]
		rwLock.Unlock()

		// 同步redis的用户缓存到Mysql
		var totalFavorited, favoriteCount, followCount, followerCount, workCount int64
		for _, userID := range backupUserIDs {
			txPipeline := rdb.TxPipeline()
			userIDStr := strconv.FormatInt(userID, 10)
			totalFavorited, _ = strconv.ParseInt(txPipeline.Get(context.Background(), getRedisKey(KeyUserTotalFavoritedPF+userIDStr)).Val(), 10, 64)
			favoriteCount, _ = strconv.ParseInt(txPipeline.Get(context.Background(), getRedisKey(KeyUserFavoriteCountPF+userIDStr)).Val(), 10, 64)
			followCount, _ = strconv.ParseInt(txPipeline.Get(context.Background(), getRedisKey(KeyUserFollowCountPF+userIDStr)).Val(), 10, 64)
			followerCount, _ = strconv.ParseInt(txPipeline.Get(context.Background(), getRedisKey(KeyUserFollowerCountPF+userIDStr)).Val(), 10, 64)
			workCount, _ = strconv.ParseInt(txPipeline.Get(context.Background(), getRedisKey(KeyUserWorkCountPF+userIDStr)).Val(), 10, 64)
			if _, err := txPipeline.Exec(context.Background()); err != nil {
				hlog.Error("dao.syncRedisToMySQL: 查询redis用户缓存失败")
				continue
			}
			if err := db.Table("users").Where("id = ?", userID).Updates(map[string]interface{}{
				"total_favorited": totalFavorited, "favorite_count": favoriteCount, "follow_count": followCount,
				"follower_count": followerCount, "work_count": workCount}).Error; err != nil {
				hlog.Error("dao.syncRedisToMySQL: 同步redis用户缓存到mysql失败")
				continue
			}
		}

		// 同步redis中的视频缓存
		var videoFavoriteCount, videoCommentCount int64
		for _, videoID := range backupVideoIDs {
			txPipeline := rdb.TxPipeline()
			videoIDStr := strconv.FormatInt(videoID, 10)
			videoFavoriteCount, _ = strconv.ParseInt(txPipeline.Get(context.Background(), getRedisKey(KeyVideoFavoriteCountPF+videoIDStr)).Val(), 10, 64)
			videoCommentCount, _ = strconv.ParseInt(txPipeline.Get(context.Background(), getRedisKey(KeyVideoCommentCountPF+videoIDStr)).Val(), 10, 64)
			if _, err := txPipeline.Exec(context.Background()); err != nil {
				hlog.Error("dao.syncRedisToMySQL: 查询redis视频缓存失败")
				continue
			}
			if err := db.Table("videos").Where("id = ?", videoID).Updates(map[string]interface{}{
				"favorite_count": videoFavoriteCount, "comment_count": videoCommentCount}).Error; err != nil {
				hlog.Error("dao.syncRedisToMySQL: 同步redis视频缓存到mysql失败")
				continue
			}
		}
	}
}

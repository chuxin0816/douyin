package dal

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"douyin/src/common/utils"
	"douyin/src/config"
	"douyin/src/dal/model"
	"douyin/src/dal/query"

	"github.com/allegro/bigcache/v3"
	"github.com/bits-and-blooms/bloom/v3"
	nebula "github.com/vesoft-inc/nebula-go/v3"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/sharding"
)

const (
	ExpireTime     = time.Minute * 30
	DelayTime      = 150 * time.Millisecond
	randFactor     = 30
	advisoryLockID = 12345
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
	db           *gorm.DB
	RDB          *redis.ClusterClient
	IncrByScript *redis.Script
	sessionPool  *nebula.SessionPool
	Cache        *bigcache.BigCache
	G            = &singleflight.Group{}
	bloomFilter  *bloom.BloomFilter
	err          error
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
	// 初始化PostgreSQL
	InitPostgreSQL()

	// 初始化Redis
	InitRedis()

	// 初始化Nebula
	InitNebula()

	// 初始化BigCache
	InitBigCache()

	// 初始化布隆过滤器
	bloomFilter = bloom.NewWithEstimates(100000, 0.001)
	if err := loadDataToBloom(); err != nil {
		panic(err)
	}
}

func InitPostgreSQL() {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
		config.Conf.PostgreSQLMaster.Host,
		config.Conf.PostgreSQLMaster.Port,
		config.Conf.PostgreSQLMaster.User,
		config.Conf.PostgreSQLMaster.Password,
		config.Conf.PostgreSQLMaster.DBName,
	)

	// 连接主库
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// 连接从库
	replicas := make([]gorm.Dialector, len(config.Conf.DatabaseConfig.PostgreSQLSlaves))
	for i, slave := range config.Conf.DatabaseConfig.PostgreSQLSlaves {
		replicas[i] = postgres.Open(fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
			slave.Host,
			slave.Port,
			slave.User,
			slave.Password,
			slave.DBName,
		))
	}
	err = db.Use(dbresolver.Register(dbresolver.Config{
		Replicas: replicas,
		Policy:   dbresolver.RandomPolicy{},
	}))
	if err != nil {
		panic(err)
	}

	// 连接池
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetConnMaxIdleTime(time.Minute * 30)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetMaxIdleConns(100)
	sqlDB.SetMaxOpenConns(500)

	// 分表路由
	err = db.Use(sharding.Register(
		sharding.Config{
			ShardingKey: "convert_id",
			ShardingAlgorithm: func(columnValue any) (suffix string, err error) {
				t, ok := columnValue.(string)
				if !ok {
					return "", nil
				}
				h := utils.NewDefaultHasher()
				id := h.Sum64(t)%128 + 1
				return fmt.Sprintf("_%03d", id), nil
			},
		}, model.Message{},
	))
	if err != nil {
		panic(err)
	}

	// 链路追踪插件
	if err := db.Use(tracing.NewPlugin(tracing.WithoutMetrics())); err != nil {
		panic(err)
	}

	q = query.Use(db)
	qComment = q.Comment
	qFavorite = q.Favorite
	qUser = q.User
	qUserLogin = q.UserLogin
	qVideo = q.Video

	generateMessageTables()
}

func InitRedis() {
	RDB = redis.NewFailoverClusterClient(&redis.FailoverOptions{
		MasterName:       config.Conf.DatabaseConfig.Redis.MasterName,
		SentinelAddrs:    config.Conf.DatabaseConfig.Redis.SentinelAddrs,
		SentinelPassword: config.Conf.DatabaseConfig.Redis.SentinelPassword,
		Password:         config.Conf.DatabaseConfig.Redis.Password,
		DB:               config.Conf.DatabaseConfig.Redis.DB,
	})
	if err := RDB.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}

	if err := redisotel.InstrumentTracing(RDB); err != nil {
		panic(err)
	}

	IncrByScript = redis.NewScript(`
	if redis.call('EXISTS', KEYS[1]) == 1 then
        return redis.call('INCRBY', KEYS[1], ARGV[1])
    else
        return nil
    end
	`)
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

func InitBigCache() {
	Cache, err = bigcache.New(context.Background(), bigcache.Config{
		Shards:             1024,
		LifeWindow:         5 * time.Second,
		CleanWindow:        5 * time.Second,
		MaxEntriesInWindow: 1024 * 600,
		MaxEntrySize:       500,
		StatsEnabled:       false,
		Verbose:            true,
		HardMaxCacheSize:   1024,
	})
	if err != nil {
		panic(err)
	}
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

// 创建消息分表
func generateMessageTables() {
	// 尝试获取顾问锁，12345 为锁的唯一 ID
	var lockAcquired bool
	if err := db.Debug().Raw("SELECT pg_try_advisory_lock(?)", advisoryLockID).Scan(&lockAcquired).Error; err != nil {
		panic(err)
	}

	if !lockAcquired {
		return // 没有获取到锁，直接退出函数
	}

	// 确保在函数返回时释放顾问锁
	defer func() {
		if err := db.Exec("SELECT pg_advisory_unlock(12345)").Error; err != nil {
			panic(err)
		}
	}()

	// 开始创建表
	for i := 1; i <= 128; i++ {
		table := fmt.Sprintf("message_%03d", i)

		// 创建表
		createTableSQL := fmt.Sprintf(`
			 CREATE TABLE IF NOT EXISTS %s (
				 id bigint NOT NULL,
				 from_user_id bigint NOT NULL DEFAULT 0,
				 to_user_id bigint NOT NULL DEFAULT 0,
				 convert_id varchar NOT NULL DEFAULT '',
				 content varchar NOT NULL DEFAULT '',
				 create_time bigint NOT NULL DEFAULT 0,
				 PRIMARY KEY (id)
			 );`, table)

		if err := db.Exec(createTableSQL).Error; err != nil {
			panic(err)
		}

		// 创建索引
		createIndexSQL := fmt.Sprintf(`
			 CREATE INDEX IF NOT EXISTS idx_convertId_createTime_%s
			 ON %s (convert_id, create_time);`, table, table)

		if err := db.Exec(createIndexSQL).Error; err != nil {
			panic(err)
		}

		// 添加列注释
		commentColumns := []struct {
			column  string
			comment string
		}{
			{"from_user_id", "发送者ID"},
			{"to_user_id", "接收者ID"},
			{"convert_id", "会话ID"},
			{"content", "消息内容"},
			{"create_time", "创建时间"},
		}

		for _, col := range commentColumns {
			commentSQL := fmt.Sprintf(`
				 COMMENT ON COLUMN %s.%s IS '%s';`, table, col.column, col.comment)

			if err := db.Exec(commentSQL).Error; err != nil {
				panic(err)
			}
		}
	}
}

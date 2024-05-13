package config

import (
	"reflect"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

const (
	consulEndpoint   = "consul:8500"
	consulConfigPath = "config"
)

var (
	Conf                = &Config{}
	NoticeJwt           = make(chan struct{})
	NoticeSnowflake     = make(chan struct{})
	NoticeOss           = make(chan struct{})
	NoticeLog           = make(chan struct{})
	NoticeMySQL         = make(chan struct{})
	NoticeRedis         = make(chan struct{})
	NoticeMongo         = make(chan struct{})
	NoticeConsul        = make(chan struct{})
	NoticeKafka         = make(chan struct{})
	NoticeOpenTelemetry = make(chan struct{})
)

type Config struct {
	JwtKey               string `mapstructure:"jwt_key"`
	*SnowflakeConfig     `mapstructure:"snowflake"`
	*OssConfig           `mapstructure:"oss"`
	*HertzConfig         `mapstructure:"hertz"`
	*LogConfig           `mapstructure:"log"`
	*DatabaseConfig      `mapstructure:"database"`
	*ConsulConfig        `mapstructure:"consul"`
	*KafkaConfig         `mapstructure:"kafka"`
	*OpenTelemetryConfig `mapstructure:"open_telemetry"`
}

type SnowflakeConfig struct {
	StartTime string `mapstructure:"start_time"`
	MachineID int64  `mapstructure:"machine_id"`
}

type OssConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyId     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	BucketName      string `mapstructure:"bucket_name"`
}

type HertzConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type LogConfig struct {
	Path       string `mapstructure:"path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

type DatabaseConfig struct {
	*MySQLConfig `mapstructure:"mysql"`
	*RedisConfig `mapstructure:"redis"`
	*MongoConfig `mapstructure:"mongo"`
}

type MySQLConfig struct {
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	DBName   string `mapstructure:"dbname"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type MongoConfig struct {
	Host   string `mapstructure:"host"`
	Port   int    `mapstructure:"port"`
	DBName string `mapstructure:"dbname"`
}

type ConsulConfig struct {
	ConsulAddr   string `mapstructure:"consul_addr"`
	FeedAddr     string `mapstructure:"feed_addr"`
	UserAddr     string `mapstructure:"user_addr"`
	FavoriteAddr string `mapstructure:"favorite_addr"`
	CommentAddr  string `mapstructure:"comment_addr"`
	PublishAddr  string `mapstructure:"publish_addr"`
	RelationAddr string `mapstructure:"relation_addr"`
	MessageAddr  string `mapstructure:"message_addr"`
}

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
}

type OpenTelemetryConfig struct {
	ApiName      string `mapstructure:"api_name"`
	FeedName     string `mapstructure:"feed_name"`
	UserName     string `mapstructure:"user_name"`
	FavoriteName string `mapstructure:"favorite_name"`
	CommentName  string `mapstructure:"comment_name"`
	PublishName  string `mapstructure:"publish_name"`
	RelationName string `mapstructure:"relation_name"`
	MessageName  string `mapstructure:"message_name"`
	JaegerAddr   string `mapstructure:"jaeger_addr"`
}

func Init() {
	if err := viper.AddRemoteProvider("consul", consulEndpoint, consulConfigPath); err != nil {
		panic(err)
	}
	viper.SetConfigType("yaml")

	if err := viper.ReadRemoteConfig(); err != nil {
		panic(err)
	}

	// 把读取到的配置信息反序列化到Conf变量中
	if err := viper.Unmarshal(Conf); err != nil {
		panic(err)
	}

	// 监控配置文件变化
	go func() {
		for {
			time.Sleep(time.Second * 5)

			if err := viper.WatchRemoteConfig(); err != nil {
				klog.Error("watch remote config failed, err:%v", err)
				continue
			}

			newConf := &Config{}
			if err := viper.Unmarshal(newConf); err != nil {
				klog.Error("viper unmarshal failed, err:%v", err)
				continue
			}
			if reflect.DeepEqual(Conf, newConf) {
				continue
			}

			if !reflect.DeepEqual(Conf.JwtKey, newConf.JwtKey) {
				Conf.JwtKey = newConf.JwtKey
				NoticeJwt <- struct{}{}
			}

			if !reflect.DeepEqual(Conf.SnowflakeConfig, newConf.SnowflakeConfig) {
				Conf.SnowflakeConfig = newConf.SnowflakeConfig
				NoticeSnowflake <- struct{}{}
			}

			if !reflect.DeepEqual(Conf.OssConfig, newConf.OssConfig) {
				Conf.OssConfig = newConf.OssConfig
				NoticeOss <- struct{}{}
			}

			if !reflect.DeepEqual(Conf.LogConfig, newConf.LogConfig) {
				Conf.LogConfig = newConf.LogConfig
				NoticeLog <- struct{}{}
			}

			if !reflect.DeepEqual(Conf.DatabaseConfig, newConf.DatabaseConfig) {
				if !reflect.DeepEqual(Conf.DatabaseConfig.MySQLConfig, newConf.DatabaseConfig.MySQLConfig) {
					Conf.DatabaseConfig.MySQLConfig = newConf.DatabaseConfig.MySQLConfig
					NoticeMySQL <- struct{}{}
				}
				if !reflect.DeepEqual(Conf.DatabaseConfig.RedisConfig, newConf.DatabaseConfig.RedisConfig) {
					Conf.DatabaseConfig.RedisConfig = newConf.DatabaseConfig.RedisConfig
					NoticeRedis <- struct{}{}
				}
				if !reflect.DeepEqual(Conf.DatabaseConfig.MongoConfig, newConf.DatabaseConfig.MongoConfig) {
					Conf.DatabaseConfig.MongoConfig = newConf.DatabaseConfig.MongoConfig
					NoticeMongo <- struct{}{}
				}
			}

			if !reflect.DeepEqual(Conf.ConsulConfig, newConf.ConsulConfig) {
				Conf.ConsulConfig = newConf.ConsulConfig
				NoticeConsul <- struct{}{}
			}

			if !reflect.DeepEqual(Conf.KafkaConfig, newConf.KafkaConfig) {
				Conf.KafkaConfig = newConf.KafkaConfig
				NoticeKafka <- struct{}{}
			}

			if !reflect.DeepEqual(Conf.OpenTelemetryConfig, newConf.OpenTelemetryConfig) {
				Conf.OpenTelemetryConfig = newConf.OpenTelemetryConfig
				NoticeOpenTelemetry <- struct{}{}
			}
		}
	}()
}
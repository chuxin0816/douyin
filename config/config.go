package config

import (
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

const (
	consulEndpoint   = "consul:8500"
	consulConfigPath = "config"
)

var Conf = &Config{}

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
	*MysqlConfig `mapstructure:"mysql"`
	*RedisConfig `mapstructure:"redis"`
	*MongoConfig `mapstructure:"mongo"`
}

type MysqlConfig struct {
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
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
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
	err := viper.AddRemoteProvider("consul", consulEndpoint, consulConfigPath)
	if err != nil {
		panic(err)
	}
	viper.SetConfigType("yaml")

	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}

	// 把读取到的配置信息反序列化到Conf变量中
	if err := viper.Unmarshal(Conf); err != nil {
		panic(err)
	}

	// 监控配置文件变化
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		if err := viper.Unmarshal(Conf); err != nil {
			klog.Error("viper unmarshal failed, err:%v", err)
		} else {
			klog.Notice("config file changed")
		}
	})
}

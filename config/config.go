package config

import (
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
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

type ConsulConfig struct {
	ConsulAddr          string `mapstructure:"consul_addr"`
	FeedAddr            string `mapstructure:"feed_addr"`
	UserAddr            string `mapstructure:"user_addr"`
	FavoriteAddr        string `mapstructure:"favorite_addr"`
	CommentAddr         string `mapstructure:"comment_addr"`
	PublishAddr         string `mapstructure:"publish_addr"`
	RelationAddr        string `mapstructure:"relation_addr"`
	MessageAddr         string `mapstructure:"message_addr"`
	FeedServiceName     string `mapstructure:"feed_service_name"`
	UserServiceName     string `mapstructure:"user_service_name"`
	FavoriteServiceName string `mapstructure:"favorite_service_name"`
	CommentServiceName  string `mapstructure:"comment_service_name"`
	PublishServiceName  string `mapstructure:"publish_service_name"`
	RelationServiceName string `mapstructure:"relation_service_name"`
	MessageServiceName  string `mapstructure:"message_service_name"`
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
	//指定配置文件类型(专门用于解析远程配置文件）
	// viper.SetConfigType("json")

	viper.SetConfigName("config")   //指定配置文件的文件名称(不需要扩展名)
	viper.AddConfigPath("./config") //指定查找配置文件的路径(这里使用相对路径)

	err := viper.ReadInConfig()
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

package config

import (
	"reflect"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/kitex-contrib/config-consul/consul"
	"gopkg.in/yaml.v3"
)

const (
	consulEndpoint   = "consul:8500"
	consulConfigPath = "conf/config.yaml"
)

var (
	Conf                *Config
	NoticeJwt           = make(chan struct{})
	NoticeSnowflake     = make(chan struct{})
	NoticeOss           = make(chan struct{})
	NoticeLog           = make(chan struct{})
	NoticeMySQL         = make(chan struct{})
	NoticeRedis         = make(chan struct{})
	NoticeConsul        = make(chan struct{})
	NoticeKafka         = make(chan struct{})
	NoticeOpenTelemetry = make(chan struct{})
)

type Config struct {
	JwtKey               string `yaml:"jwt_key"`
	*SnowflakeConfig     `yaml:"snowflake"`
	*OssConfig           `yaml:"oss"`
	*HertzConfig         `yaml:"hertz"`
	*LogConfig           `yaml:"log"`
	*DatabaseConfig      `yaml:"database"`
	*ConsulConfig        `yaml:"consul"`
	*KafkaConfig         `yaml:"kafka"`
	*OpenTelemetryConfig `yaml:"open_telemetry"`
}

type SnowflakeConfig struct {
	StartTime string `yaml:"start_time"`
	MachineID int64  `yaml:"machine_id"`
}

type OssConfig struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyId     string `yaml:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret"`
	BucketName      string `yaml:"bucket_name"`
}

type HertzConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type LogConfig struct {
	Path       string `yaml:"path"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
}

type DatabaseConfig struct {
	MySQLMaster *MySQLConfig   `yaml:"mysql-master"`
	MySQLSlaves []*MySQLConfig `yaml:"mysql-slaves"`
	Redis       *RedisConfig   `yaml:"redis"`
	Nebula      *NebulaConfig  `yaml:"nebula"`
}

type MySQLConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	DBName   string `yaml:"dbname"`
}

type RedisConfig struct {
	MasterName       string   `yaml:"master_name"`
	SentinelAddrs    []string `yaml:"sentinel_addrs"`
	SentinelPassword string   `yaml:"sentinel_password"`
	Password         string   `yaml:"password"`
	DB               int      `yaml:"db"`
}

type NebulaConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Space    string `yaml:"space"`
}

type ConsulConfig struct {
	ConsulAddr   string `yaml:"consul_addr"`
	UserAddr     string `yaml:"user_addr"`
	VideoAddr    string `yaml:"video_addr"`
	FavoriteAddr string `yaml:"favorite_addr"`
	CommentAddr  string `yaml:"comment_addr"`
	RelationAddr string `yaml:"relation_addr"`
	MessageAddr  string `yaml:"message_addr"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
}

type OpenTelemetryConfig struct {
	ApiName      string `yaml:"api_name"`
	UserName     string `yaml:"user_name"`
	VideoName    string `yaml:"video_name"`
	FavoriteName string `yaml:"favorite_name"`
	CommentName  string `yaml:"comment_name"`
	RelationName string `yaml:"relation_name"`
	MessageName  string `yaml:"message_name"`
	MetricAddr   string `yaml:"metric_addr"`
	JaegerAddr   string `yaml:"jaeger_addr"`
}

func Init() {
	client, err := consul.NewClient(consul.Options{
		Addr: consulEndpoint,
	})
	if err != nil {
		panic(err)
	}

	// 监听配置变化
	client.RegisterConfigCallback(consulConfigPath, consul.AllocateUniqueID(), func(s string, cp consul.ConfigParser) {
		newConf := &Config{}
		if err := yaml.Unmarshal([]byte(s), newConf); err != nil {
			klog.Error("config unmarshal failed, err:%v", err)
		}

		if Conf == nil {
			Conf = newConf
			return
		}

		if reflect.DeepEqual(Conf, newConf) {
			return
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
			if !reflect.DeepEqual(Conf.DatabaseConfig.MySQLMaster, newConf.DatabaseConfig.MySQLMaster) {
				Conf.DatabaseConfig.MySQLMaster = newConf.DatabaseConfig.MySQLMaster
				NoticeMySQL <- struct{}{}
			}
			if !reflect.DeepEqual(Conf.DatabaseConfig.MySQLSlaves, newConf.DatabaseConfig.MySQLSlaves) {
				Conf.DatabaseConfig.MySQLSlaves = newConf.DatabaseConfig.MySQLSlaves
				NoticeMySQL <- struct{}{}
			}
			if !reflect.DeepEqual(Conf.DatabaseConfig.Redis, newConf.DatabaseConfig.Redis) {
				Conf.DatabaseConfig.Redis = newConf.DatabaseConfig.Redis
				NoticeRedis <- struct{}{}
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
	})
}

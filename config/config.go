package config

import (
	"errors"
	"log"
	"time"

	"github.com/spf13/viper"
)


type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	MongoDB  MongoDB
	Cookie   Cookie
	Store    Store
	Session  Session
	Metrics  Metrics
	Logger   Logger
	AWS      AWS
	Jaeger   Jaeger
}


type ServerConfig struct {
	AppVersion        string
	Port              string
	PprofPort         string
	Mode              string
	JwtSecretKey      string
	CookieName        string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	SSL               bool
	CtxDefaultTimeout time.Duration
	CSRF              bool
	Debug             bool
}

type Logger struct {
	Development       bool
	DisableCaller     bool
	DisableStacktrace bool
	Encoding          string
	Level             string
}


type PostgresConfig struct {
	PostgresqlHost     string
	PostgresqlPort     string
	PostgresqlUser     string
	PostgresqlPassword string
	PostgresqlDbname   string
	PostgresqlSSLMode  bool
	PgDriver           string
}

type RedisConfig struct {
	RedisAddr      string
	RedisPassword  string
	RedisDB        string
	RedisDefaultdb string
	MinIdleConns   int
	PoolSize       int
	PoolTimeout    int
	Password       string
	DB             int
}


type MongoDB struct {
	MongoURI string
}


type Cookie struct {
	Name     string
	MaxAge   int
	Secure   bool
	HTTPOnly bool
}


type Session struct {
	Prefix string
	Name   string
	Expire int
}


type Metrics struct {
	URL         string
	ServiceName string
}


type Store struct {
	ImagesFolder string
}


type AWS struct {
	Endpoint       string
	MinioAccessKey string
	MinioSecretKey string
	UseSSL         bool
	MinioEndpoint  string
}


type Jaeger struct {
	Host        string
	ServiceName string
	LogSpans    bool
}


func LoadConfig(filename string) (*viper.Viper, error) {
	v := viper.New()

	v.SetConfigName(filename)
	v.AddConfigPath(".")
	v.AutomaticEnv()
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, errors.New("config file not found")
		}
		return nil, err
	}

	return v, nil
}


func ParseConfig(v *viper.Viper) (*Config, error) {
	var c Config

	err := v.Unmarshal(&c)
	if err != nil {
		log.Printf("unable to decode into struct, %v", err)
		return nil, err
	}

	return &c, nil
}

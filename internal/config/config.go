package config

import (
	"github.com/Jagerente/gocfg"
	"github.com/Jagerente/gocfg/pkg/values"
	"gorm.io/gorm/logger"

	"time"
)

type App struct {
	Local        bool     `env:"LOCAL" default:"true"`
	LocalOrigins []string `env:"LOCAL_ORIGINS,omitempty" default:"http://localhost:*" description:"List of allowed origins for CORS"`
}

type LoggerConfig struct {
	LogLevel int8 `env:"LOG_LEVEL" default:"-1" description:"https://pkg.go.dev/github.com/rs/zerolog@v1.33.0#Level"`
}

type PostgresConfig struct {
	Host           string          `env:"POSTGRES_HOST" default:"postgres:5432"`
	Username       string          `env:"POSTGRES_USERNAME" default:"postgres"`
	Password       string          `env:"POSTGRES_PASSWORD" default:"12345"`
	Database       string          `env:"POSTGRES_DATABASE" default:"workshop"`
	Params         string          `env:"POSTGRES_PARAMS,omitempty"`
	MigrationsPath string          `env:"POSTGRES_MIGRATIONS_PATH,omitempty" example:"/migrations/postgres"`
	DoMigrate      bool            `env:"POSTGRES_DO_MIGRATE" default:"false" description:"Whether to run app-driver migrations on start"`
	LogLevel       logger.LogLevel `env:"POSTGRES_LOGLEVEL" default:"3" description:"https://pkg.go.dev/gorm.io/gorm/logger@v1.31.0#LogLevel"`
}

type RouterConfig struct {
	ServerPort uint16 `env:"SERVER_PORT" default:"8080"`
	Debug      bool   `env:"ROUTER_DEBUG" default:"true"`
	AuthDebug  bool   `env:"AUTH_DEBUG" default:"true"`
}

type RedisConfig struct {
	Host     string `env:"REDIS_HOST" default:"redis:6379"`
	DB       int    `env:"REDIS_DB" default:"0"`
	Username string `env:"REDIS_USERNAME,omitempty"`
	Password string `env:"REDIS_PASSWORD,omitempty"`
}

type Config struct {
	App         App              `title:"App configuration"`
	Logger      LoggerConfig     `title:"Logger configuration"`
	Router      RouterConfig     `title:"Router configuration"`
	RedisConfig RedisConfig      `title:"Redis configuration"`
	Postgres    PostgresConfig   `title:"Postgres configuration"`
	RateLimits  RateLimitsConfig `title:"Rate limits configuration"`
}

type RateLimitsConfig struct {
	PostsLimit     uint64        `env:"POSTS_LIMIT" default:"30"`
	PostsPeriod    time.Duration `env:"POSTS_PERIOD" default:"1h" description:"https://pkg.go.dev/time#ParseDuration"`
	CommentsLimit  uint64        `env:"COMMENTS_LIMIT" default:"100"`
	CommentsPeriod time.Duration `env:"COMMENTS_PERIOD" default:"30m" description:"https://pkg.go.dev/time#ParseDuration"`
}

func New() (*Config, error) {
	var cfg = new(Config)

	cfgManager := gocfg.NewDefault()
	if dotEnvProvider, err := values.NewDotEnvProvider(); err == nil {
		cfgManager = cfgManager.AddValueProviders(dotEnvProvider)
	}

	if err := cfgManager.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

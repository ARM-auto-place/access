package config

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ARM-auto-place/access/internal/metrics"
	"github.com/caarlos0/env/v10"
	"github.com/rs/zerolog"
)

type PostgresConfig struct {
	Addr            string        `env:"PG_ADDR"`
	DB              string        `env:"PG_DB"`
	User            string        `env:"PG_USER"`
	Password        string        `env:"PG_PASSWORD"`
	SSLMode         string        `env:"PG_SSL_MODE" envDefault:"require"`
	MaxConn         int           `env:"PG_MAX_CONN" envDefault:"10"`
	MaxIdleLifetime time.Duration `env:"PG_MAX_IDLE_LIFETIME" envDefault:"30s"`
	MaxLifetime     time.Duration `env:"PG_MAX_LIFETIME" envDefault:"3m"`
	ConnectTimeout  time.Duration `env:"PG_CONNECT_TIMEOUT" envDefault:"5s"`
	PrepareCacheCap int           `env:"PG_PREPARE_CACHE_CAP" envDefault:"128"`
	ReplicaAddr     string        `env:"PG_REPLICA_ADDR"`
	UserRO          string        `env:"PG_REPLICA_USER"`
	PasswordRO      string        `env:"PG_REPLICA_PASSWORD"`
}

type CacheConfig struct {
	RefreshMaxDirtyUsers int           `env:"CACHE_REFRESH_MAX_DIRTY_USERS" envDefault:"100000"`
	MaxCost              int64         `env:"CACHE_MAX_COST" envDefault:"100000"`
	NumCounters          int64         `env:"CACHE_NUM_COUNTERS" envDefault:"1000000"`
	BufferItems          int64         `env:"CACHE_BUFFER_ITEMS" envDefault:"64"`
	TTL                  time.Duration `env:"CACHE_TTL" envDefault:"5m"`
	RefreshInterval      time.Duration `env:"CACHE_REFRESH_INTERVAL" envDefault:"1m"`
	RefreshBatchSize     int           `env:"CACHE_REFRESH_BATCH_SIZE" envDefault:"1000"`
	RefreshTimeout       time.Duration `env:"CACHE_REFRESH_TIMEOUT" envDefault:"30s"`
}

type Config struct {
	ServiceName string `env:"SERVICE_NAME" envDefault:"access"`

	LogLevel  string `env:"LOGGER_LEVEL" envDefault:"debug"`
	LogFormat string `env:"LOGGER_FORMAT" envDefault:""`

	ServiceBind    string   `env:"BIND_ADDR" envDefault:":9000"`
	HealthBind     string   `env:"BIND_HEALTH" envDefault:":9091"`
	MetricsBind    string   `env:"BIND_METRICS" envDefault:":9090"`
	MetricsPath    string   `env:"METRICS_PATH" envDefault:"/metrics"`
	AllowedOrigins []string `env:"ALLOWED_ORIGINS" envSeparator:","`

	MaxRequestBodySize   int           `env:"MAX_REQUEST_BODY_SIZE" envDefault:"104857600"` // 100 MB
	MaxRequestHeaderSize int           `env:"MAX_REQUEST_HEADER_SIZE" envDefault:"16384"`   // 16 KB
	ReadTimeout          time.Duration `env:"READ_TIMEOUT" envDefault:"120s"`
	WriteTimeout         time.Duration `env:"WRITE_TIMEOUT" envDefault:"30s"`
	ShutdownTimeout      time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`

	Postgres PostgresConfig
	Cache    CacheConfig
}

var (
	configuration *Config
	configMu      sync.Once
)

func internalConfig() *Config {
	configMu.Do(func() {
		configuration = &Config{}
		if err := env.Parse(configuration); err != nil {
			panic(err)
		}

		if configuration.Postgres.ReplicaAddr == "" {
			configuration.Postgres.ReplicaAddr = configuration.Postgres.Addr
		}
		if configuration.Postgres.UserRO == "" {
			configuration.Postgres.UserRO = configuration.Postgres.User
			configuration.Postgres.PasswordRO = configuration.Postgres.Password
		}

		for i, origin := range configuration.AllowedOrigins {
			configuration.AllowedOrigins[i] = strings.TrimSpace(origin)
		}

	})

	return configuration
}

func Values() Config {
	return *internalConfig()
}

func (cfg Config) Logger() zerolog.Logger {
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil || cfg.LogLevel == "" {
		level = zerolog.DebugLevel
	}
	logger := zerolog.New(os.Stdout)
	if cfg.LogFormat == "console" {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}
	return logger.Level(level).With().Timestamp().Str("service", cfg.ServiceName).Logger()
}

var (
	metricsValue *metrics.Metrics
	metricsMu    sync.Once
)

func Metrics() *metrics.Metrics {
	metricsMu.Do(func() {
		metricsValue = metrics.NewMetrics(internalConfig().ServiceName)
	})
	return metricsValue
}

package config

import (
	"os"
	"strconv"
	"time"
	"github.com/shopspring/decimal"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	NATS      NATSConfig
	Vault     VaultConfig
	PayPal    PayPalConfig
	Stripe    StripeConfig
	Proxy     ProxyConfig
	Risk      RiskConfig
	Scheduler SchedulerConfig
}

type ServerConfig struct {
	Port           int           `yaml:"port" default:"8080"`
	GRPCPort       int           `yaml:"grpc_port" default:"9090"`
	ReadTimeout    time.Duration `yaml:"read_timeout" default:"30s"`
	WriteTimeout   time.Duration `yaml:"write_timeout" default:"30s"`
	JWTSecret      string        `yaml:"jwt_secret"`
	AdminJWTExpiry time.Duration `yaml:"admin_jwt_expiry" default:"24h"`
	Env            string        `yaml:"env" default:"development"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host" default:"localhost"`
	Port     int    `yaml:"port" default:"5432"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname" default:"ab_payment"`
	SSLMode  string `yaml:"ssl_mode" default:"disable"`
	MaxConns int    `yaml:"max_conns" default:"50"`
}

type RedisConfig struct {
	Addrs    []string `yaml:"addrs" default:"localhost:6379"`
	Password string   `yaml:"password"`
	DB       int      `yaml:"db" default:"0"`
	PoolSize int      `yaml:"pool_size" default:"20"`
}

type NATSConfig struct {
	URLs  []string `yaml:"urls" default:"nats://localhost:4222"`
	Token string   `yaml:"token"`
}

type VaultConfig struct {
	Address string `yaml:"address" default:"http://localhost:8200"`
	Token   string `yaml:"token"`
	Path    string `yaml:"path" default:"secret/ab-payment"`
}

type PayPalConfig struct {
	IsSandbox bool   `yaml:"is_sandbox" default:"true"`
	RetryMax  int    `yaml:"retry_max" default:"3"`
}

type StripeConfig struct {
	RetryMax int `yaml:"retry_max" default:"3"`
}

type ProxyConfig struct {
	Providers     []ProxyProvider `yaml:"providers"`
	DefaultType   string          `yaml:"default_type" default:"residential"`
	HealthCheckInterval time.Duration `yaml:"health_check_interval" default:"30s"`
	MaxFailCount  int             `yaml:"max_fail_count" default:"3"`
}

type ProxyProvider struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Endpoint string `yaml:"endpoint"`
	APIKey   string `yaml:"api_key"`
	PoolSize int    `yaml:"pool_size" default:"10"`
}

type RiskConfig struct {
	EnableMLDetection   bool          `yaml:"enable_ml_detection" default:"false"`
	DefaultThrottleTTL  time.Duration `yaml:"default_throttle_ttl" default:"30m"`
	MaxConsecutiveFails int           `yaml:"max_consecutive_fails" default:"3"`
	MinSuccessRate      float64       `yaml:"min_success_rate" default:"0.7"`
}

type SchedulerConfig struct {
	DefaultStrategy     string          `yaml:"default_strategy" default:"weighted_round_robin"`
	AllocationTimeout   time.Duration   `yaml:"allocation_timeout" default:"5s"`
	AccountCacheTTL     time.Duration   `yaml:"account_cache_ttl" default:"30s"`
	CircuitHalfOpenWait time.Duration   `yaml:"circuit_half_open_wait" default:"60s"`
	DefaultSingleMin    decimal.Decimal `yaml:"default_single_min"`
	DefaultSingleMax    decimal.Decimal `yaml:"default_single_max"`
	DefaultDailyMax     decimal.Decimal `yaml:"default_daily_max"`
	DefaultMonthlyMax   decimal.Decimal `yaml:"default_monthly_max"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{}
	// In production, use viper to load from file
	// For now, load from environment with defaults
	cfg.Server.Port = envInt("SERVER_PORT", 8080)
	cfg.Server.GRPCPort = envInt("GRPC_PORT", 9090)
	cfg.Server.JWTSecret = envStr("JWT_SECRET", "change-me-in-production")
	cfg.Server.AdminJWTExpiry = envDuration("ADMIN_JWT_EXPIRY", 24*time.Hour)
	cfg.Server.Env = envStr("ENV", "development")

	cfg.Database.Host = envStr("DB_HOST", "localhost")
	cfg.Database.Port = envInt("DB_PORT", 5432)
	cfg.Database.User = envStr("DB_USER", "postgres")
	cfg.Database.Password = envStr("DB_PASSWORD", "")
	cfg.Database.DBName = envStr("DB_NAME", "ab_payment")
	cfg.Database.SSLMode = envStr("DB_SSLMODE", "disable")
	cfg.Database.MaxConns = envInt("DB_MAX_CONNS", 50)

	cfg.Redis.Addrs = []string{envStr("REDIS_ADDR", "localhost:6379")}
	cfg.Redis.Password = envStr("REDIS_PASSWORD", "")
	cfg.Redis.DB = envInt("REDIS_DB", 0)
	cfg.Redis.PoolSize = envInt("REDIS_POOL_SIZE", 20)

	cfg.NATS.URLs = []string{envStr("NATS_URL", "nats://localhost:4222")}

	cfg.Proxy.DefaultType = envStr("PROXY_DEFAULT_TYPE", "residential")
	cfg.Proxy.MaxFailCount = envInt("PROXY_MAX_FAIL_COUNT", 3)

	cfg.Risk.MaxConsecutiveFails = envInt("RISK_MAX_FAILS", 3)
	cfg.Risk.MinSuccessRate = envFloat("RISK_MIN_SUCCESS_RATE", 0.7)

	cfg.Scheduler.DefaultStrategy = envStr("SCHEDULER_STRATEGY", "weighted_round_robin")

	return cfg, nil
}

func envStr(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func envInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

func envFloat(key string, defaultVal float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return defaultVal
}

func envDuration(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultVal
}


package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the feedback service
type Config struct {
	Server   ServerConfig
	MongoDB  MongoDBConfig
	Redis    RedisConfig
	Auth     AuthConfig
	Storage  StorageConfig
	Notifier NotifierConfig
	Comment  CommentConfig
	Feedback FeedbackConfig
	Logging  LoggingConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port            int
	Host            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// MongoDBConfig holds MongoDB configuration
type MongoDBConfig struct {
	URI             string
	Database        string
	MaxPoolSize     uint64
	MinPoolSize     uint64
	MaxConnIdleTime time.Duration
}

// RedisConfig holds Redis configuration for caching
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// AuthConfig holds auth service configuration
type AuthConfig struct {
	ServiceURL   string
	BaseURL      string
	ClientID     string
	ClientSecret string
	CacheSeconds int
	Timeout      int
	SkipPaths    []string
}

// StorageConfig holds storage service configuration
type StorageConfig struct {
	ServiceURL   string
	MaxFileSize  int64
	AllowedTypes []string
}

// NotifierConfig holds notifier service configuration
type NotifierConfig struct {
	ServiceURL   string
	ClientID     string
	ClientSecret string
	AdminUserID  string
	Enabled      bool
}

// CommentConfig holds comment service configuration
type CommentConfig struct {
	ServiceURL string
	Timeout    int
}

// FeedbackConfig holds feedback-specific configuration
type FeedbackConfig struct {
	RequireApproval        bool
	AllowAnonymous         bool
	MaxTitleLength         int
	MaxDescriptionLength   int
	MaxAttachments         int
	TrendingWeightVotes    float64
	TrendingWeightComments float64
	TrendingWeightViews    float64
	TrendingDecayHours     int
	RateLimitRequests      int
	RateLimitWindow        int
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level    string
	Format   string
	FilePath string
	Encoding string
	Logger   string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	_ = godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Port:            getEnvAsInt("SERVER_PORT", 5012),
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:     getDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:    getDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:     getDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
			ShutdownTimeout: getDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		MongoDB: MongoDBConfig{
			URI:             getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database:        getEnv("MONGODB_DATABASE", "minisource_feedback"),
			MaxPoolSize:     uint64(getEnvAsInt("MONGODB_MAX_POOL_SIZE", 100)),
			MinPoolSize:     uint64(getEnvAsInt("MONGODB_MIN_POOL_SIZE", 10)),
			MaxConnIdleTime: getDuration("MONGODB_MAX_CONN_IDLE_TIME", 10*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Auth: AuthConfig{
			ServiceURL:   getEnv("AUTH_SERVICE_URL", "http://localhost:8080"),
			BaseURL:      getEnv("AUTH_BASE_URL", "http://localhost:8080"),
			ClientID:     getEnv("AUTH_CLIENT_ID", "feedback-service"),
			ClientSecret: getEnv("AUTH_CLIENT_SECRET", "feedback-service-secret-key"),
			CacheSeconds: getEnvAsInt("AUTH_CACHE_SECONDS", 300),
			Timeout:      getEnvAsInt("AUTH_TIMEOUT", 30),
			SkipPaths:    getEnvAsSlice("AUTH_SKIP_PATHS", []string{"/health", "/ready", "/live"}),
		},
		Storage: StorageConfig{
			ServiceURL:   getEnv("STORAGE_SERVICE_URL", "http://localhost:5001"),
			MaxFileSize:  int64(getEnvAsInt("STORAGE_MAX_FILE_SIZE", 10485760)), // 10MB
			AllowedTypes: getEnvAsSlice("STORAGE_ALLOWED_TYPES", []string{"image/jpeg", "image/png", "image/gif", "image/webp", "application/pdf"}),
		},
		Notifier: NotifierConfig{
			ServiceURL:   getEnv("NOTIFIER_SERVICE_URL", "http://localhost:9002"),
			ClientID:     getEnv("NOTIFIER_CLIENT_ID", "feedback-service"),
			ClientSecret: getEnv("NOTIFIER_CLIENT_SECRET", "feedback-service-secret-key"),
			AdminUserID:  getEnv("NOTIFIER_ADMIN_USER_ID", ""),
			Enabled:      getEnvAsBool("NOTIFIER_ENABLED", true),
		},
		Comment: CommentConfig{
			ServiceURL: getEnv("COMMENT_SERVICE_URL", "http://localhost:5010"),
			Timeout:    getEnvAsInt("COMMENT_TIMEOUT", 30),
		},
		Feedback: FeedbackConfig{
			RequireApproval:        getEnvAsBool("FEEDBACK_REQUIRE_APPROVAL", false),
			AllowAnonymous:         getEnvAsBool("FEEDBACK_ALLOW_ANONYMOUS", false),
			MaxTitleLength:         getEnvAsInt("FEEDBACK_MAX_TITLE_LENGTH", 200),
			MaxDescriptionLength:   getEnvAsInt("FEEDBACK_MAX_DESCRIPTION_LENGTH", 10000),
			MaxAttachments:         getEnvAsInt("FEEDBACK_MAX_ATTACHMENTS", 5),
			TrendingWeightVotes:    getEnvAsFloat("FEEDBACK_TRENDING_WEIGHT_VOTES", 1.0),
			TrendingWeightComments: getEnvAsFloat("FEEDBACK_TRENDING_WEIGHT_COMMENTS", 2.0),
			TrendingWeightViews:    getEnvAsFloat("FEEDBACK_TRENDING_WEIGHT_VIEWS", 0.1),
			TrendingDecayHours:     getEnvAsInt("FEEDBACK_TRENDING_DECAY_HOURS", 72),
			RateLimitRequests:      getEnvAsInt("FEEDBACK_RATE_LIMIT_REQUESTS", 100),
			RateLimitWindow:        getEnvAsInt("FEEDBACK_RATE_LIMIT_WINDOW", 60),
		},
		Logging: LoggingConfig{
			Level:    getEnv("LOG_LEVEL", "info"),
			Format:   getEnv("LOG_FORMAT", "json"),
			FilePath: getEnv("LOG_FILE_PATH", "./logs/feedback.log"),
			Encoding: getEnv("LOG_ENCODING", "json"),
			Logger:   getEnv("LOG_LOGGER", "zap"),
		},
	}, nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

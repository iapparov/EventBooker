package config

import (
	"fmt"
	"os"
	"time"

	wbfconfig "github.com/wb-go/wbf/config"
)

// AppConfig is the root configuration for the application.
type AppConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
	DB       DBConfig       `mapstructure:"db_config"`
	Telegram TelegramConfig `mapstructure:"telegram"`
	Mail     MailConfig     `mapstructure:"mail"`
	Retry    RetryConfig    `mapstructure:"retry_strategy"`
	Gin      GinConfig      `mapstructure:"gin"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	User     UserConfig     `mapstructure:"username_config"`
	Password PasswordConfig `mapstructure:"password_config"`
	Event    EventConfig    `mapstructure:"event_config"`
}

type RetryConfig struct {
	Attempts int           `mapstructure:"attempts" default:"3"`
	Delay    time.Duration `mapstructure:"delay" default:"1s"`
	Backoffs float64       `mapstructure:"backoffs" default:"2"`
}

type GinConfig struct {
	Mode string `mapstructure:"mode" default:"debug"`
}

type ServerConfig struct {
	Host string `mapstructure:"host" default:"localhost"`
	Port int    `mapstructure:"port" default:"8080"`
}

type LoggerConfig struct {
	Level string `mapstructure:"level" default:"info"`
}

type RabbitMQConfig struct {
	Host              string `mapstructure:"host" default:"localhost"`
	Port              int    `mapstructure:"port" default:"5672"`
	User              string
	Password          string
	ConnectionTimeout int    `mapstructure:"connection_timeout"`
	Heartbeat         int    `mapstructure:"heartbeat"`
	ConnectionName    string `mapstructure:"connection_name"`
	Exchange          string `mapstructure:"exchange"`
	QueueName         string `mapstructure:"queue_name"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string
	Password string
	DBName   string `mapstructure:"db_name"`
	SSLMode  string `mapstructure:"ssl_mode" default:"disable"`
}

type DBConfig struct {
	Master          PostgresConfig   `mapstructure:"postgres"`
	Slaves          []PostgresConfig `mapstructure:"slaves"`
	MaxOpenConns    int              `mapstructure:"max_open_conns"`
	MaxIdleConns    int              `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration    `mapstructure:"conn_max_lifetime"`
	QueryTimeout    time.Duration    `mapstructure:"query_timeout" default:"5s"`
}

type TelegramConfig struct {
	BotToken string
}

type MailConfig struct {
	SMTPHost     string `mapstructure:"smtp_host" default:""`
	SMTPPort     int    `mapstructure:"smtp_port" default:"587"`
	SMTPEmail    string `mapstructure:"smtp_user" default:""`
	SMTPPassword string `mapstructure:"smtp_password" default:""`
}

type JWTConfig struct {
	ExpAccessToken  int `mapstructure:"jwt_exp_access_token"`
	ExpRefreshToken int `mapstructure:"jwt_exp_refresh_token"`
	AccessSecret    string
	RefreshSecret   string
}

type UserConfig struct {
	MinLength         int    `mapstructure:"min_length"`
	MaxLength         int    `mapstructure:"max_length"`
	AllowedCharacters string `mapstructure:"allowed_characters"`
	CaseInsensitive   bool   `mapstructure:"case_insensitive"`
}

type PasswordConfig struct {
	MinLength    int  `mapstructure:"min_length"`
	MaxLength    int  `mapstructure:"max_length"`
	RequireUpper bool `mapstructure:"require_upper"`
	RequireLower bool `mapstructure:"require_lower"`
	RequireDigit bool `mapstructure:"require_digit"`
}

type EventConfig struct {
	NameMinLength        int          `mapstructure:"name_min_length"`
	NameMaxLength        int          `mapstructure:"name_max_length"`
	DescriptionMaxLength int          `mapstructure:"desctiption_max_length"`
	DescriptionRequired  bool         `mapstructure:"description_require"`
	TTL                  []int        `mapstructure:"booking_ttl"`
	SupportedTTLs        map[int]bool `mapstructure:"-"`
}

// MustLoad loads configuration from files and environment variables.
func MustLoad() *AppConfig {
	cfg := wbfconfig.New()

	if err := cfg.LoadEnvFiles("./.env"); err != nil {
		panic(fmt.Sprintf("failed to load env files: %v", err))
	}

	cfg.EnableEnv("")

	if err := cfg.LoadConfigFiles("./config/local.yaml"); err != nil {
		panic(fmt.Sprintf("failed to load config files: %v", err))
	}

	var appCfg AppConfig
	if err := cfg.Unmarshal(&appCfg); err != nil {
		panic(fmt.Sprintf("failed to unmarshal config: %v", err))
	}

	appCfg.DB.Master.DBName = os.Getenv("POSTGRES_DB")
	appCfg.DB.Master.User = os.Getenv("POSTGRES_USER")
	appCfg.DB.Master.Password = os.Getenv("POSTGRES_PASSWORD")

	appCfg.RabbitMQ.User = os.Getenv("RABBITMQ_USER")
	appCfg.RabbitMQ.Password = os.Getenv("RABBITMQ_PASSWORD")

	appCfg.Telegram.BotToken = os.Getenv("TELEGRAM_BOT_TOKEN")

	appCfg.Mail.SMTPEmail = os.Getenv("MAIL_SMTP_USER")
	appCfg.Mail.SMTPPassword = os.Getenv("MAIL_SMTP_PASSWORD")

	appCfg.JWT.AccessSecret = os.Getenv("JWT_ACCESS_SECRET")
	appCfg.JWT.RefreshSecret = os.Getenv("JWT_REFRESH_SECRET")

	appCfg.Event.SupportedTTLs = buildSupportedTTLs(appCfg.Event.TTL)

	return &appCfg
}

func buildSupportedTTLs(formats []int) map[int]bool {
	m := make(map[int]bool, len(formats))
	for _, f := range formats {
		m[f] = true
	}
	return m
}

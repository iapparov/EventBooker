package config

import (
	"fmt"
	wbfconfig "github.com/wb-go/wbf/config"
	"os"
)

func NewAppConfig() (*AppConfig, error) {
	envFilePath := "./.env"
	appConfigFilePath := "./config/local.yaml"

	cfg := wbfconfig.New()

	// Загрузка .env файлов
	if err := cfg.LoadEnvFiles(envFilePath); err != nil {
		return nil, fmt.Errorf("failed to load env files: %w", err)
	}

	// Включение поддержки переменных окружения
	cfg.EnableEnv("")

	if err := cfg.LoadConfigFiles(appConfigFilePath); err != nil {
		return nil, fmt.Errorf("failed to load config files: %w", err)
	}

	var appCfg AppConfig
	if err := cfg.Unmarshal(&appCfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	appCfg.DBConfig.Master.DBName = os.Getenv("POSTGRES_DB")
	appCfg.DBConfig.Master.User = os.Getenv("POSTGRES_USER")
	appCfg.DBConfig.Master.Password = os.Getenv("POSTGRES_PASSWORD")

	appCfg.RabbitmqConfig.User = os.Getenv("RABBITMQ_USER")
	appCfg.RabbitmqConfig.Password = os.Getenv("RABBITMQ_PASSWORD")

	appCfg.TelegramConfig.BotToken = os.Getenv("TELEGRAM_BOT_TOKEN")

	appCfg.MailConfig.SMTPEmail = os.Getenv("MAIL_SMTP_USER")
	appCfg.MailConfig.SMTPPassword = os.Getenv("MAIL_SMTP_PASSWORD")

	appCfg.JwtConfig.JwtAccessSecret = os.Getenv("JWT_ACCESS_SECRET")
	appCfg.JwtConfig.JwtRefreshSecret = os.Getenv("JWT_REFRESH_SECRET")
	appCfg.EventConfig.SupportedTTLs = configTTLS(appCfg.EventConfig.TTL)
	return &appCfg, nil
}

func configTTLS(formats []int) map[int]bool {
	suppTTL := make(map[int]bool, len(formats))
	for _, f := range formats {
		suppTTL[f] = true
	}
	return suppTTL
}

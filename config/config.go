package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App AppConfig
	DB  DBConfig
	JWT JWTConfig
}

type AppConfig struct {
	Port string
	Env  string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret    string
	AccessTTL time.Duration
}

// Load đọc config từ file .env (nếu có) và biến môi trường.
// Biến môi trường luôn có độ ưu tiên cao hơn file .env.
func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	viper.SetDefault("PORT", "3636")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "123123")
	viper.SetDefault("DB_NAME", "todo_api_golang")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("JWT_SECRET", "dev-secret-change-me")
	viper.SetDefault("JWT_ACCESS_TTL_MINUTES", 15)

	// Bỏ qua lỗi nếu không tìm thấy file .env — biến môi trường vẫn được đọc.
	_ = viper.ReadInConfig()

	return &Config{
		App: AppConfig{
			Port: viper.GetString("PORT"),
			Env:  viper.GetString("APP_ENV"),
		},
		DB: DBConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			Name:     viper.GetString("DB_NAME"),
			SSLMode:  viper.GetString("DB_SSLMODE"),
		},
		JWT: JWTConfig{
			Secret:    viper.GetString("JWT_SECRET"),
			AccessTTL: time.Duration(viper.GetInt("JWT_ACCESS_TTL_MINUTES")) * time.Minute,
		},
	}, nil
}

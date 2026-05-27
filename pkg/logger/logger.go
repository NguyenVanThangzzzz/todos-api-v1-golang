package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger là alias để gói code dễ đọc - tương đương Winston/Pino bên Node
type Logger = zap.Logger

// New khởi tạo logger với output dễ đọc cho development (màu, level chữ).
// Production nên dùng zap.NewProduction() để xuất JSON cho log aggregator (ELK, Loki...).
func New() *Logger {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return l
}

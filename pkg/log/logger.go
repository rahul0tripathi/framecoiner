package log

import "go.uber.org/zap"

type Logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
}

func NewZapLogger(dev bool, options ...zap.Option) (Logger, error) {
	if dev {
		return zap.NewDevelopment(options...)
	}

	return zap.NewProduction(options...)
}

package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

func NewLogger(config Config) (*Logger, error) {
	zaplvl := zap.NewAtomicLevel()
	if err := zaplvl.UnmarshalText([]byte(config.Level)); err != nil {
		return nil, err
	}

	zapconfig := zap.NewDevelopmentEncoderConfig()
	zapconfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zapEncoder := zapcore.NewConsoleEncoder(zapconfig)

	core := zapcore.NewCore(zapEncoder, zapcore.AddSync(os.Stdout), zaplvl)

	zapLogger := zap.New(core, zap.AddCaller())

	return &Logger{
		Logger: zapLogger,
	}, nil
}

func NewTestLogger() *Logger {
	return &Logger{
		Logger: zap.NewNop(),
	}
}

func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(fields...),
	}
}

type ctxKey string

const logKey ctxKey = "log"

func ContextWithLogger(ctx context.Context, log *Logger) context.Context {
	return context.WithValue(ctx, logKey, log)
}

func FromContext(ctx context.Context) *Logger {
	return ctx.Value(logKey).(*Logger)
}

package logger

import (
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

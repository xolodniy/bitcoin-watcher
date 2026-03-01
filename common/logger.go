package common

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

var self *zap.SugaredLogger

func Logger() *zap.SugaredLogger {
	if self == nil {
		return zap.NewExample().Sugar()
	}
	return self
}

func InitLogger(fNotify func(string), level string) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zap.DebugLevel
	case "info":
		zapLevel = zap.InfoLevel
	case "warn", "warning":
		zapLevel = zap.WarnLevel
	case "error":
		zapLevel = zap.ErrorLevel
	case "fatal":
		zapLevel = zap.FatalLevel
	case "panic":
		zapLevel = zap.PanicLevel
	default:
		fmt.Println("unexpected logger level, used info level by default")
		zapLevel = zap.InfoLevel
	}
	loggerConfig := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapLevel),
		Development: true,
		Encoding:    "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:  "time",
			LevelKey: "level",
			NameKey:  "logger",
			// CallerKey:     "caller", // FIXME: Adjust it properly
			MessageKey:    "msg",
			StacktraceKey: "stacktrace",
			LineEnding:    zapcore.DefaultLineEnding,
			// EncodeLevel:   zapcore.CapitalLevelEncoder,
			EncodeLevel: zapcore.CapitalColorLevelEncoder,
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Format("15:04:05")) // Format without milliseconds and timezone
				// enc.AppendString(t.Format("2006-01-02 15:04:05")) // Format without milliseconds and timezone
			},
			EncodeDuration:   zapcore.StringDurationEncoder,
			ConsoleSeparator: " ",
			// EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	logger, err := loggerConfig.Build()
	if err != nil {
		panic(err)
	}

	// хук для отправки ошибок в тг
	hook := func(entry zapcore.Entry) error {
		if entry.Level >= zapcore.WarnLevel {
			fNotify(fmt.Sprintf("[%s] %s", entry.Level.String(), entry.Message))
		}
		return nil
	}

	// оборачиваем core хуком через WrapCore
	logger = logger.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.RegisterHooks(core, hook)
	}))

	// теперь можно сахар
	self = logger.Sugar()
}

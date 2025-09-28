package logger

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const serviceName = "notifier"

var (
	globalLogger *zap.SugaredLogger
	initOnce     sync.Once
	initErr      error
)

func Init(level zapcore.Level) error {
	initOnce.Do(func() {
		config := zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
		config.OutputPaths = []string{"stdout"}
		config.ErrorOutputPaths = []string{"stdout"}
		config.Level.SetLevel(level)
		config.EncoderConfig.StacktraceKey = "stacktrace"

		l, err := config.Build(zap.AddCallerSkip(1))
		if err != nil {
			initErr = err
			return
		}

		globalLogger = l.With(zap.String("service", serviceName)).Sugar()
	})

	return initErr
}

type LogCtxKey string

const loggerCtxKey LogCtxKey = "logger"

func ToContext(ctx context.Context, l *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, loggerCtxKey, l)
}

func With(args ...interface{}) *zap.SugaredLogger {
	return globalLogger.With(args...)
}

func Debugf(ctx context.Context, msg string, keysAndValues ...interface{}) {
	if l, ok := ctx.Value(loggerCtxKey).(*zap.SugaredLogger); ok && l != nil {
		l.Debugf(msg, keysAndValues...)
	}

	globalLogger.Debugf(msg, keysAndValues...)
}

func Infof(ctx context.Context, msg string, keysAndValues ...interface{}) {
	if l, ok := ctx.Value(loggerCtxKey).(*zap.SugaredLogger); ok && l != nil {
		l.Infof(msg, keysAndValues...)
		return
	}

	globalLogger.Infof(msg, keysAndValues...)
}

func Warnf(ctx context.Context, msg string, keysAndValues ...interface{}) {
	if l, ok := ctx.Value(loggerCtxKey).(*zap.SugaredLogger); ok && l != nil {
		l.Warnf(msg, keysAndValues...)
		return
	}

	globalLogger.Warnf(msg, keysAndValues...)
}

func Errorf(ctx context.Context, msg string, keysAndValues ...interface{}) {
	if l, ok := ctx.Value(loggerCtxKey).(*zap.SugaredLogger); ok && l != nil {
		l.Errorf(msg, keysAndValues...)
	}

	globalLogger.Errorf(msg, keysAndValues...)
}

func Fatalf(ctx context.Context, msg string, keysAndValues ...interface{}) {
	if l, ok := ctx.Value(loggerCtxKey).(*zap.SugaredLogger); ok && l != nil {
		l.Fatalf(msg, keysAndValues...)
	}

	globalLogger.Fatalf(msg, keysAndValues...)
}

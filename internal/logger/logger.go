package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)

	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})

	Sync() error
}

type loggerImpl struct {
	base    *zap.Logger
	sugared *zap.SugaredLogger
}

func New(level string, pretty bool) Logger {
	var cfg zap.Config
	if pretty {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg = zap.NewProductionConfig()
	}

	if lvl := parseLevel(level); lvl != nil {
		cfg.Level = zap.NewAtomicLevelAt(*lvl)
	}

	base, err := cfg.Build(
		zap.AddStacktrace(zapcore.FatalLevel), // Only add stack traces for Fatal
	)
	if err != nil {
		panic(err)
	}

	return &loggerImpl{
		base:    base,
		sugared: base.Sugar(),
	}
}

func parseLevel(lvl string) *zapcore.Level {
	switch lvl {
	case "debug":
		l := zapcore.DebugLevel
		return &l
	case "info":
		l := zapcore.InfoLevel
		return &l
	case "warn":
		l := zapcore.WarnLevel
		return &l
	case "error":
		l := zapcore.ErrorLevel
		return &l
	default:
		return nil
	}
}

func (l *loggerImpl) Debug(msg string, fields ...zap.Field) { l.base.Debug(msg, fields...) }
func (l *loggerImpl) Info(msg string, fields ...zap.Field)  { l.base.Info(msg, fields...) }
func (l *loggerImpl) Warn(msg string, fields ...zap.Field)  { l.base.Warn(msg, fields...) }
func (l *loggerImpl) Error(msg string, fields ...zap.Field) { l.base.Error(msg, fields...) }
func (l *loggerImpl) Fatal(msg string, fields ...zap.Field) { l.base.Fatal(msg, fields...) }

func (l *loggerImpl) Debugf(t string, args ...interface{}) { l.sugared.Debugf(t, args...) }
func (l *loggerImpl) Infof(t string, args ...interface{})  { l.sugared.Infof(t, args...) }
func (l *loggerImpl) Warnf(t string, args ...interface{})  { l.sugared.Warnf(t, args...) }
func (l *loggerImpl) Errorf(t string, args ...interface{}) { l.sugared.Errorf(t, args...) }
func (l *loggerImpl) Fatalf(t string, args ...interface{}) { l.sugared.Fatalf(t, args...) }

func (l *loggerImpl) Sync() error { return l.base.Sync() }

// Field constructors (re-exported from zap for convenience)
// This allows other packages to use structured logging without importing zap directly.
func String(key, val string) zap.Field                 { return zap.String(key, val) }
func Int(key string, val int) zap.Field                { return zap.Int(key, val) }
func Duration(key string, val time.Duration) zap.Field { return zap.Duration(key, val) }
func Error(err error) zap.Field                        { return zap.Error(err) }

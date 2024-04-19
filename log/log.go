package log

import (
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

type Config struct {
	Format   string
	Level    string
	FilePath string
	Rotate   RotateConfig
}

type RotateConfig struct {
	Compress   bool
	MaxSize    int
	MaxAge     int
	MaxBackups int
}

func Init(config *Config) {
	logger = NewSugaredLogger(config)
}

func Debug(msg string) {
	logger.Debug(msg)
	logger.Named("debug").Debug(msg)
}

func Debugf(msg string, args ...interface{}) {
	logger.Debugf(msg, args...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	logger.Debugw(msg, keysAndValues...)
}

func Info(msg string) {
	logger.Info(msg)
}

func Infof(msg string, args ...interface{}) {
	logger.Infof(msg, args...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	logger.Infow(msg, keysAndValues...)
}

func Warn(msg string) {
	logger.Warn(msg)
}

func Warnf(msg string, args ...interface{}) {
	logger.Warnf(msg, args...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	logger.Warnw(msg, keysAndValues...)
}

func Error(msg string) {
	logger.Error(msg)
}

func Errorf(msg string, args ...interface{}) {
	logger.Errorf(msg, args...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	logger.Errorw(msg, keysAndValues...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args)
}

func Fatalf(template string, args ...interface{}) {
	logger.Fatalf(template, args...)
}

func Fatalw(msg string, keysAndValues ...interface{}) {
	logger.Fatalw(msg, keysAndValues...)
}

func Panic(args ...interface{}) {
	logger.Panic(args)
}

func Panicf(template string, args ...interface{}) {
	logger.Panicf(template, args...)
}

func Panicw(msg string, keysAndValues ...interface{}) {
	logger.Panicw(msg, keysAndValues...)
}

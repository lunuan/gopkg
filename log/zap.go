package log

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	DebugLevel = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel

	RotateMaxSize    = 300
	RotateMaxAge     = 7
	RotateMaxBackups = 15
)

func NewLogger(conf *Config) *zap.Logger {
	return initZapLogger(conf)
}

func NewSugaredLogger(conf *Config) *zap.SugaredLogger {
	log := initZapLogger(conf)
	return log.Sugar()
}

func initZapLogger(conf *Config) *zap.Logger {
	//log rotate
	var rotateHook *lumberjack.Logger
	if conf.FilePath != "" {
		maxSize := RotateMaxSize
		if conf.Rotate.MaxSize > 100 {
			maxSize = conf.Rotate.MaxSize
		}

		maxAge := RotateMaxAge
		if conf.Rotate.MaxAge >= 3 {
			maxAge = conf.Rotate.MaxAge
		}

		maxBackups := RotateMaxBackups
		if conf.Rotate.MaxBackups >= 3 {
			maxBackups = conf.Rotate.MaxBackups
		}

		rotateHook = &lumberjack.Logger{
			Filename:   conf.FilePath,
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge,
			Compress:   conf.Rotate.Compress,
			LocalTime:  true,
		}
	}

	//log encoder
	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05,000")
	// encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeName = zapcore.FullNameEncoder
	encoderConfig.ConsoleSeparator = " "
	encoderConfig.LineEnding = zapcore.DefaultLineEnding
	encoderConfig.MessageKey = "message"
	encoderConfig.StacktraceKey = "stacktrace"
	encoderConfig.CallerKey = "caller"
	encoderConfig.FunctionKey = "function"

	switch conf.Format {
	case "json":
		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case "console":
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	case "common":
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = NewCommonEncoder(encoderConfig)
	default:
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = NewkvEncoder(encoderConfig)
	}

	//log Level
	logLevel := zap.NewAtomicLevel()
	logLevel.SetLevel(toZapLevel(conf.Level))

	//log fileWrites consoleWrites
	var fileWrites zapcore.WriteSyncer
	var consoleWrites zapcore.WriteSyncer
	var core zapcore.Core

	consoleWrites = zapcore.AddSync(os.Stdout)
	if rotateHook != nil {
		fileWrites = zapcore.AddSync(rotateHook)
		fileCore := zapcore.NewCore(encoder, fileWrites, logLevel)
		consoleCore := zapcore.NewCore(encoder, consoleWrites, logLevel)
		core = zapcore.NewTee(fileCore, consoleCore)
	} else {
		core = zapcore.NewCore(encoder, consoleWrites, logLevel)
	}

	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return zapLogger
}

func toZapLevel(level string) zapcore.Level {
	level = strings.ToLower(level)
	logLevel := zap.NewAtomicLevel()
	err := logLevel.UnmarshalText([]byte(level))
	if err != nil {
		logLevel.SetLevel(zapcore.InfoLevel)
	}
	return logLevel.Level()
}

package config

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogConfig ...
type LogConfig struct {
	Level           string   `mapstructure:"level"`
	Mode            string   `mapstructure:"mode"`
	Encoding        string   `mapstructure:"encoding"`
	StacktraceLevel string   `mapstructure:"stacktrace_level"`
	MaskedFields    []string `mapstructure:"masked_fields"`
}

const (
	production  = "production"
	development = "development"
	json        = "json"
	console     = "console"
)

var levelMap = map[string]zapcore.Level{
	"debug":  zap.DebugLevel,
	"info":   zap.InfoLevel,
	"warn":   zap.WarnLevel,
	"error":  zap.ErrorLevel,
	"dpanic": zap.DPanicLevel,
	"panic":  zap.PanicLevel,
	"fatal":  zap.FatalLevel,
}

func validateLogConfig(conf LogConfig) {
	_, existed := levelMap[conf.Level]
	if !existed {
		panic("log level must be one of: debug, info, warn, error, dpanic, panic, fatal")
	}
	if conf.Mode != development && conf.Mode != production {
		panic("log mode must be one of: development, production")
	}
	if conf.Encoding != console && conf.Encoding != json {
		panic("log encoding must be one of: console, json")
	}

	if conf.StacktraceLevel != "" {
		_, existed := levelMap[conf.StacktraceLevel]
		if !existed {
			panic("invalid stacktrace_level")
		}
	}
}

// NewLogger creates a logger
func NewLogger(conf LogConfig) *zap.Logger {
	validateLogConfig(conf)

	zapConf := zap.NewProductionConfig()
	if conf.Mode == development {
		zapConf = zap.NewDevelopmentConfig()
	}

	level := zap.NewAtomicLevelAt(levelMap[conf.Level])
	zapConf.Level = level
	zapConf.Encoding = conf.Encoding

	var options []zap.Option
	if conf.StacktraceLevel != "" {
		stackLevel := zap.NewAtomicLevelAt(levelMap[conf.StacktraceLevel])
		options = append(options, zap.AddStacktrace(stackLevel))
	}
	options = append(options, zap.AddCaller())

	logger, err := zapConf.Build(options...)
	if err != nil {
		panic(err)
	}
	return logger
}

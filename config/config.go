package config

import (
	"erpc/plugins"
	"erpc/plugins/etcd"
	"erpc/plugins/logger"
	etcdv3 "github.com/coreos/etcd/clientv3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

//后面用适配器模式实现,注册中心配置对象
type RegistryConfig struct {
	RegistryType string   `yaml:"registry_type"` //etcd default
	Addr         []string `yaml:"endpoints"`
	UserName     string   `yaml:"user_name"`
	Pass         string   `yaml:"pass"`
}

func (conf *RegistryConfig) NewRegistry() (registry plugins.Registry, err error) {
	cli, err := etcdv3.New(etcdv3.Config{
		Endpoints: conf.Addr,
		Username:  conf.UserName,
		Password:  conf.Pass,
	})
	return etcd.NewEtcdRegistry(cli), nil
}

//日志组件配置对象
type LoggerConfig struct {
	Level      string `yaml:"level"`       //debug  info  warn  error
	Encoding   string `yaml:"encoding"`    //json or console
	CallFull   bool   `yaml:"call_full"`   //whether full call path or short path, default is short
	Filename   string `yaml:"file_name"`   //log file name
	MaxSize    int    `yaml:"max_size"`    //max size of log.(MB)
	MaxAge     int    `yaml:"max_age"`     //time to keep, (day)
	MaxBackups int    `yaml:"max_backups"` //max file numbers
	LocalTime  bool   `yaml:"local_time"`  //(default UTC)
	Compress   bool   `yaml:"compress"`    //default false
}

func (l *LoggerConfig) NewLogger() *logger.ZapLogger {
	if l.Filename == "" {
		logObj, _ := zap.NewProduction(zap.AddCallerSkip(2))
		return logger.NewZapLogger(logObj)
	}

	enCfg := zap.NewProductionEncoderConfig()
	if l.CallFull {
		enCfg.EncodeCaller = zapcore.FullCallerEncoder
	}
	encoder := zapcore.NewJSONEncoder(enCfg)
	if l.Encoding == "console" {
		zapcore.NewConsoleEncoder(enCfg)
	}

	//zapWriter := zapcore.
	zapWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   l.Filename,
		MaxSize:    l.MaxSize,
		MaxAge:     l.MaxAge,
		MaxBackups: l.MaxBackups,
		LocalTime:  l.LocalTime,
	})

	newCore := zapcore.NewCore(encoder, zapWriter, zap.NewAtomicLevelAt(convertLogLevel(l.Level)))
	opts := []zap.Option{zap.ErrorOutput(zapWriter)}
	opts = append(opts, zap.AddCaller(), zap.AddCallerSkip(2))
	logObj := zap.New(newCore, opts...)
	return logger.NewZapLogger(logObj)
}

func convertLogLevel(levelStr string) (level zapcore.Level) {
	switch levelStr {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	}
	return
}

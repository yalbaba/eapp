package zap

import (
	"eapp/logger"
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"time"
)

type DefaultLogger struct {
	logger *zap.SugaredLogger
}

func NewDefaultLogger(conf *logger.LoggerConfig) logger.ILogger {
	//构建自定义日志组件（这里采用zap）
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		//EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
		EncodeLevel: zapcore.CapitalColorLevelEncoder, //这里可以指定颜色
		//EncodeTime:     zapcore.ISO8601TimeEncoder,       // ISO8601 UTC 时间格式
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		EncodeDuration: zapcore.SecondsDurationEncoder,
		//EncodeCaller:   zapcore.ShortCallerEncoder, // 全路径编码器(会显示zap包的运行日志)
	}

	// 设置日志级别
	atom := zap.NewAtomicLevelAt(zap.InfoLevel)
	config := zap.Config{
		Level:         atom,          // 日志级别
		Development:   true,          // 开发模式，堆栈跟踪
		Encoding:      "console",     // 输出格式 console 或 json
		EncoderConfig: encoderConfig, // 编码器配置
		//InitialFields:    map[string]interface{}{"serviceName": "wisdom_park"}, // 这里可以标识服务的名称，消息的尾巴上会打印这里的key和value
		OutputPaths:      []string{"stdout"}, // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
		ErrorOutputPaths: []string{"stderr"},
	}
	config.EncoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder //这里可以指定颜色
	// 构建日志
	var err error
	l, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("log 初始化失败: %v", err))
	}

	return &DefaultLogger{
		logger: l.Sugar(),
	}
}

func getWriter(filename, outPutDir string) io.Writer {
	// 生成rotatelogs的Logger 实际生成的文件名 demo.log.YYmmddHH
	// demo.log是指向最新日志的链接
	// 保存7天内的日志，每1小时(整点)分割一次日志
	hook, err := rotatelogs.New(
		outPutDir+filename+".%Y%m%d",
		rotatelogs.WithLinkName(filename),
		rotatelogs.WithMaxAge(time.Hour*24*7),
		rotatelogs.WithRotationTime(time.Hour*24),
	)
	if err != nil {
		panic(err)
	}
	return hook
}

func (zl *DefaultLogger) Info(args ...interface{}) {
	zl.logger.Info(args...)
}

func (zl *DefaultLogger) Infof(format string, args ...interface{}) {
	zl.logger.Infof(format, args...)
}

func (zl *DefaultLogger) Debug(args ...interface{}) {
	zl.logger.Debug(args...)
}

func (zl *DefaultLogger) Debugf(format string, args ...interface{}) {
	zl.logger.Debugf(format, args...)
}

func (zl *DefaultLogger) Warn(args ...interface{}) {
	zl.logger.Warn(args...)
}

func (zl *DefaultLogger) Warnf(format string, args ...interface{}) {
	zl.logger.Warnf(format, args...)
}

func (zl *DefaultLogger) Error(args ...interface{}) {
	zl.logger.Error(args...)
}

func (zl *DefaultLogger) Errorf(format string, args ...interface{}) {
	zl.logger.Errorf(format, args...)
}

func (zl *DefaultLogger) DPanic(args ...interface{}) {
	zl.logger.DPanic(args...)
}

func (zl *DefaultLogger) DPanicf(format string, args ...interface{}) {
	zl.logger.DPanicf(format, args...)
}

func (zl *DefaultLogger) Panic(args ...interface{}) {
	zl.logger.Panic(args...)
}

func (zl *DefaultLogger) Panicf(format string, args ...interface{}) {
	zl.logger.Panicf(format, args...)
}

func (zl *DefaultLogger) Fatal(args ...interface{}) {
	zl.logger.Fatal(args...)
}

func (zl *DefaultLogger) Fatalf(format string, args ...interface{}) {
	zl.logger.Fatalf(format, args...)
}

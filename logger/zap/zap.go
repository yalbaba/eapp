package zap

import (
	"erpc/logger"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"time"
)

type DefaultLogger struct {
	logger *zap.Logger
}

func NewDefaultLogger(conf *logger.LoggerConfig) logger.ILogger {
	//构建自定义日志组件（这里采用zap）
	encoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey: "msg",
		LevelKey:   "level",
		TimeKey:    "ts",
		//CallerKey:      "file",
		CallerKey:     "caller",
		StacktraceKey: "trace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		//EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		//EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(int64(d) / 1000000)
		},
	})

	// 实现两个判断日志等级的interface
	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return true
	})

	warnLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.WarnLevel
	})

	infoHook_1 := os.Stdout
	infoHook_2 := getWriter(conf.OutFile, conf.OutputDir)
	errorHook := getWriter(conf.ErrFile, conf.OutputDir)

	// 最后创建具体的Logger
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(infoHook_1), infoLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(infoHook_2), infoLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(errorHook), warnLevel),
	)

	// 需要传入 zap.AddCaller() 才会显示打日志点的文件名和行数, 有点小坑
	l := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	return &DefaultLogger{
		logger: l,
	}
}

func getWriter(filename, outPutDir string) io.Writer {
	// 生成rotatelogs的Logger 实际生成的文件名 demo.log.YYmmddHH
	// demo.log是指向最新日志的链接
	// 保存7天内的日志，每1小时(整点)分割一次日志
	hook, err := rotatelogs.New(
		// 没有使用go风格反人类的format格式
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
	zl.logger.Sugar().Info(args...)
}

func (zl *DefaultLogger) Debug(args ...interface{}) {
	zl.logger.Sugar().Info(args...)
}

func (zl *DefaultLogger) Warn(args ...interface{}) {
	zl.logger.Sugar().Info(args...)
}

func (zl *DefaultLogger) Error(args ...interface{}) {
	zl.logger.Sugar().Info(args...)
}

func (zl *DefaultLogger) DPanic(args ...interface{}) {
	zl.logger.Sugar().Info(args...)
}

func (zl *DefaultLogger) Panic(args ...interface{}) {
	zl.logger.Sugar().Info(args...)
}

func (zl *DefaultLogger) Fatal(args ...interface{}) {
	zl.logger.Sugar().Info(args...)
}

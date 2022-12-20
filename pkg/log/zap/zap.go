// Package zap copy from https://www.uoften.com/article/185822.html
package zap

import (
	"fmt"
	"io"
	"strings"
	"time"

	"path/filepath"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ahfuzhang/cheaprpc/pkg/config"
	"github.com/ahfuzhang/cheaprpc/pkg/util/debugs"
	stringsutil "github.com/ahfuzhang/cheaprpc/pkg/util/strings"
)

var errorLogger *zap.SugaredLogger
var logLevel zapcore.Level = zapcore.DebugLevel

// Init 初始化日志
func Init(app string, server string, cfg *config.Zap) error {
	var (
		level   zapcore.Level
		useJSON bool
	)
	if err := level.UnmarshalText(stringsutil.NoAllocBytes(cfg.Level)); err != nil {
		return debugs.WarpError(err, "log name error")
	}
	if cfg.Format == "text" {
		useJSON = false
	} else if cfg.Format == "json" {
		useJSON = true
	} else {
		return fmt.Errorf("[%s]log format error", debugs.SourceCodeLoc(1))
	}
	zapCfg := zapcore.EncoderConfig{
		MessageKey:  "msg",
		LevelKey:    "level",
		EncodeLevel: zapcore.CapitalLevelEncoder,
		TimeKey:     "ts",
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		CallerKey:    "file",
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(int64(d) / 1000000)
		},
	}
	var encoder zapcore.Encoder
	if useJSON {
		encoder = zapcore.NewJSONEncoder(zapCfg)
	} else {
		encoder = zapcore.NewConsoleEncoder(zapCfg)
	}
	logLevel = level
	writer, err := getWriter(filepath.Join(cfg.LogPath, fmt.Sprintf("%s_%s.log", app, server)), cfg)
	if err != nil {
		return debugs.WarpError(err, "open log fail")
	}
	core := zapcore.NewTee(zapcore.NewCore(encoder, zapcore.AddSync(writer),
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= logLevel
		})))
	log := zap.New(core, zap.AddCaller()) // 需要传入 zap.AddCaller() 才会显示打日志点的文件名和行数, 有点小坑
	errorLogger = log.Sugar()
	return nil
}

func SetLogLevel(level zapcore.Level) {
	logLevel = level
}

func getWriter(filename string, cfg *config.Zap) (io.Writer, error) {
	// 生成rotatelogs的Logger 实际生成的文件名 demo.log.YYmmddHH
	// demo.log是指向最新日志的链接
	// 保存7天内的日志，每1小时(整点)分割一次日志
	hook, err := rotatelogs.New(
		strings.Replace(filename, ".log", "", -1)+"-%Y%m%d%H.log", // 没有使用go风格反人类的format格式
		rotatelogs.WithMaxAge(time.Duration(cfg.MaxAgeMinutes)*time.Minute),
		rotatelogs.WithRotationSize(1024*1024*cfg.RotationSizeMB),
	)
	if err != nil {
		return nil, debugs.WarpError(err, "rotatelogs.New error")
	}
	return hook, nil
}
func Debug(args ...interface{}) {
	errorLogger.Debug(args...)
}
func Debugf(template string, args ...interface{}) {
	errorLogger.Debugf(template, args...)
}
func Info(args ...interface{}) {
	errorLogger.Info(args...)
}
func Infof(template string, args ...interface{}) {
	errorLogger.Infof(template, args...)
}
func Warn(args ...interface{}) {
	errorLogger.Warn(args...)
}
func Warnf(template string, args ...interface{}) {
	errorLogger.Warnf(template, args...)
}
func Error(args ...interface{}) {
	errorLogger.Error(args...)
}
func Errorf(template string, args ...interface{}) {
	errorLogger.Errorf(template, args...)
}
func DPanic(args ...interface{}) {
	errorLogger.DPanic(args...)
}
func DPanicf(template string, args ...interface{}) {
	errorLogger.DPanicf(template, args...)
}
func Panic(args ...interface{}) {
	errorLogger.Panic(args...)
}
func Panicf(template string, args ...interface{}) {
	errorLogger.Panicf(template, args...)
}
func Fatal(args ...interface{}) {
	errorLogger.Fatal(args...)
}
func Fatalf(template string, args ...interface{}) {
	errorLogger.Fatalf(template, args...)
}

// todo: 增加 WithContext()方法，把context中的关键字段打印进去

package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

var (
	Log    *zap.SugaredLogger
	xLog   *zap.SugaredLogger
	prefix string
)

func init() {
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zapcore.DebugLevel)
	syncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./logs/wallet.log",
		MaxSize:    6,
		MaxAge:     10,
		MaxBackups: 10,
		LocalTime:  true,
		Compress:   false,
	})

	zapLog := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "name",
		CallerKey:      "line",
		MessageKey:     "msg",
		FunctionKey:    "F",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}),
		zapcore.NewMultiWriteSyncer(syncer, zapcore.AddSync(os.Stdout)), atomicLevel), zap.AddCaller())

	zap.ReplaceGlobals(zapLog)

	Log = zap.S()
	xLog = Log.WithOptions(zap.AddCallerSkip(1))
}
func Debugw(msg string, keyAndValues ...interface{}) {
	xLog.Debugw(prefix+msg, keyAndValues...)
}

func Infow(msg string, keyAndValues ...interface{}) {
	xLog.Infow(prefix+msg, keyAndValues...)
}

func Warnw(msg string, keyAndValues ...interface{}) {
	xLog.Warnw(prefix+msg, keyAndValues...)
}

func Errorw(msg string, keyAndValues ...interface{}) {
	xLog.Errorw(prefix+msg, keyAndValues...)
}

func Debugf(template string, args ...interface{}) {
	xLog.Debugf(prefix+template, args...)
}

func Infof(template string, args ...interface{}) {
	xLog.Infof(prefix+template, args...)
}

func Warnf(template string, args ...interface{}) {
	xLog.Warnf(prefix+template, args...)
}

func Errorf(template string, args ...interface{}) {
	xLog.Errorf(prefix+template, args...)
}

func Debug(args ...interface{}) {
	xLog.Debug(append([]interface{}{prefix}, args...)...)
}

func Info(args ...interface{}) {
	xLog.Info(append([]interface{}{prefix}, args...)...)
}

func Warn(args ...interface{}) {
	xLog.Warn(append([]interface{}{prefix}, args...)...)
}

func Error(args ...interface{}) {
	xLog.Error(append([]interface{}{prefix}, args...)...)
}

func With(args ...interface{}) *zap.SugaredLogger {
	return xLog.With(args...)
}

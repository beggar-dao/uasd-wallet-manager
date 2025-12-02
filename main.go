package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"uasd-wallet-manager/controllers"
	"uasd-wallet-manager/log"
	"uasd-wallet-manager/routers"
)

func main() {

	err := keyInit()
	if err != nil {
		log.Errorw("read key  err", zap.Error(err))
		return
	}

	// 1. 初始化 zap + lumberjack 日志配置（你的原有配置）
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zapcore.DebugLevel) // 日志级别：Debug及以上都输出

	// 日志轮转配置（写入 ./logs/wallet.log）
	syncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./logs/wallet.log", // 日志文件路径
		MaxSize:    6,                   // 单个文件最大 6MB
		MaxAge:     10,                  // 日志保留 10 天
		MaxBackups: 10,                  // 最多保留 10 个备份文件
		LocalTime:  true,                // 文件名使用本地时间
		Compress:   false,               // 不压缩备份文件
	})

	// 2. 配置 zap 日志格式（JSON 格式，和业务日志保持一致）
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder   // 时间格式：2025-11-27T17:52:18.347Z
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // 日志级别：INFO/ERROR（大写）
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // JSON 格式输出
		syncer,                                // 输出到 lumberjack（文件轮转）
		atomicLevel,                           // 日志级别
	)

	// 3. 初始化 zap 日志实例（添加调用者信息：文件名+行号）
	logger := zap.New(core, zap.AddCaller())
	defer logger.Sync()        // 程序退出时刷新日志缓存
	zap.ReplaceGlobals(logger) // 全局替换 zap 日志，方便业务代码调用

	// 4. 关键步骤：重定向 Gin 日志到 zap 的日志同步器
	// 让 Gin 日志同时输出到：日志文件（syncer） + 控制台（os.Stdout）
	gin.DefaultWriter = io.MultiWriter(syncer, os.Stdout)

	// 5. 初始化 Gin 引擎（gin.Default() 包含默认日志中间件）
	r := gin.Default()

	routers.Adder(r)

	// 7. 启动服务（打印启动日志）
	zap.L().Info("Wallet 服务启动成功", zap.String("port", ":10832"))
	controllers.ConsolidationManagerStart()
	if err := r.Run(":10832"); err != nil {
		zap.L().Fatal("服务启动失败", zap.Error(err))
	}

}
func keyInit() error {
	err := godotenv.Load("./key.env")
	if err != nil {
		log.Errorw("Error loading .env filer", zap.Error(err))
	}
	controllers.Key = os.Getenv("key")
	log.Infof(controllers.Key)
	return nil
}

package main

import (
    "context"
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "gopkg.in/natefinch/lumberjack.v2"
    "io"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    "uasd-wallet-manager/controllers"
    "uasd-wallet-manager/log"
    "uasd-wallet-manager/routers"
)

func main() {
    err := keyInit()
    if err != nil {
        log.Errorw("read key err", zap.Error(err))
        return
    }

    // 初始化 zap + lumberjack 日志
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

    encoderConfig := zap.NewProductionEncoderConfig()
    encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
    core := zapcore.NewCore(
        zapcore.NewJSONEncoder(encoderConfig),
        syncer,
        atomicLevel,
    )

    logger := zap.New(core, zap.AddCaller())
    defer logger.Sync()
    zap.ReplaceGlobals(logger)

    // Gin 日志重定向
    gin.DefaultWriter = io.MultiWriter(syncer, os.Stdout)

    // 初始化 Gin
    r := gin.Default()
    routers.Adder(r)

    // 启动后台任务（非阻塞）
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    go controllers.ConsolidationManagerStart(ctx)

    // 启动 HTTP 服务
    srv := &http.Server{
        Addr:    ":10832",
        Handler: r,
    }
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            zap.L().Fatal("服务启动失败", zap.Error(err))
        }
    }()
    zap.L().Info("Wallet 服务启动成功", zap.String("port", ":10832"))

    // 优雅退出
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    zap.L().Info("收到退出信号，开始优雅关闭")

    cancel() // 通知后台任务退出

    ctxTimeout, _ := context.WithTimeout(context.Background(), 5*time.Second)
    if err := srv.Shutdown(ctxTimeout); err != nil {
        zap.L().Error("HTTP 优雅关闭失败", zap.Error(err))
    }
}

func keyInit() error {
    err := godotenv.Load("./key.env")
    if err != nil {
        log.Errorw("Error loading .env file", zap.Error(err))
    }
    controllers.Key = os.Getenv("key")
    log.Infof(controllers.Key)
    return nil
}

package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"polaris-io/backend/app/mqueue/cmd/job/internal/config"
	"polaris-io/backend/app/mqueue/cmd/job/internal/handler"
	"polaris-io/backend/app/mqueue/cmd/job/internal/svc"

	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f", "etc/mqueue.yaml", "Specify the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	// 设置日志
	if err := c.SetUp(); err != nil {
		panic(err)
	}

	logx.Info("Starting mqueue job service...")

	// 创建服务上下文
	svcCtx := svc.NewServiceContext(c)

	// 创建 asynq server
	server := svc.NewAsynqServer(c)

	// 创建 asynq client（用于在处理任务时入队新任务）
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     c.Redis.Addr,
		Password: c.Redis.Password,
	})
	defer asynqClient.Close()

	// 创建 mux 并注册处理器
	mux := asynq.NewServeMux()
	handler.RegisterHandlers(mux, svcCtx, asynqClient)

	// 优雅关闭
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		logx.Info("Shutting down mqueue job service...")
		server.Shutdown()
	}()

	// 启动服务
	logx.Info("Mqueue job service started, waiting for tasks...")
	if err := server.Run(mux); err != nil {
		logx.Errorf("Mqueue job service error: %v", err)
		os.Exit(1)
	}
}

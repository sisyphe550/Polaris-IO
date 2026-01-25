package main

import (
	"flag"
	"fmt"

	"polaris-io/backend/app/search/cmd/job/internal/config"
	"polaris-io/backend/app/search/cmd/job/internal/handler"
	"polaris-io/backend/app/search/cmd/job/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f", "etc/search-job.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 配置日志
	logx.MustSetup(c.Log)

	fmt.Printf("Starting search-job service...\n")

	svcCtx := svc.NewServiceContext(c)
	consumer := handler.NewKafkaConsumer(svcCtx)

	// 启动 Kafka 消费者（阻塞）
	consumer.Start()
}

package config

import "github.com/zeromicro/go-zero/core/service"

type Config struct {
	service.ServiceConf

	// Kafka 消费者配置
	Kafka struct {
		Brokers []string
		Topic   string
		GroupID string
	}

	// Elasticsearch 配置
	ES struct {
		Addresses []string
		Username  string
		Password  string
		Index     string
	}
}

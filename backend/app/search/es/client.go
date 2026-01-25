package es

import (
	"crypto/tls"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/zeromicro/go-zero/core/logx"
)

// Config ES 配置
type Config struct {
	Addresses []string // ES 地址列表
	Username  string   // 用户名
	Password  string   // 密码
	Index     string   // 索引名称
}

// Client ES 客户端封装
type Client struct {
	client *elasticsearch.Client
	index  string
}

// NewClient 创建 ES 客户端
func NewClient(cfg Config) (*Client, error) {
	esCfg := elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // 开发环境跳过证书验证
			},
		},
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, err
	}

	// 测试连接
	res, err := client.Info()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		logx.Errorf("ES connection error: %s", res.String())
		return nil, err
	}

	logx.Info("ES connected successfully")

	return &Client{
		client: client,
		index:  cfg.Index,
	}, nil
}

// GetClient 获取原始 ES 客户端
func (c *Client) GetClient() *elasticsearch.Client {
	return c.client
}

// GetIndex 获取索引名称
func (c *Client) GetIndex() string {
	return c.index
}

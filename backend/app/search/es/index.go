package es

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

// FileIndexMapping 文件索引的 mapping 定义
const FileIndexMapping = `
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0,
    "analysis": {
      "analyzer": {
        "filename_analyzer": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": ["lowercase", "edge_ngram_filter"]
        },
        "filename_search_analyzer": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": ["lowercase"]
        }
      },
      "filter": {
        "edge_ngram_filter": {
          "type": "edge_ngram",
          "min_gram": 1,
          "max_gram": 20
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "id": {
        "type": "keyword"
      },
      "user_id": {
        "type": "long"
      },
      "file_id": {
        "type": "long"
      },
      "name": {
        "type": "text",
        "analyzer": "filename_analyzer",
        "search_analyzer": "filename_search_analyzer",
        "fields": {
          "keyword": {
            "type": "keyword"
          }
        }
      },
      "name_pinyin": {
        "type": "text",
        "analyzer": "filename_analyzer",
        "search_analyzer": "filename_search_analyzer"
      },
      "ext": {
        "type": "keyword"
      },
      "size": {
        "type": "long"
      },
      "hash": {
        "type": "keyword"
      },
      "parent_id": {
        "type": "long"
      },
      "is_dir": {
        "type": "boolean"
      },
      "create_time": {
        "type": "date"
      },
      "update_time": {
        "type": "date"
      }
    }
  }
}
`

// IndexExists 检查索引是否存在
func (c *Client) IndexExists(ctx context.Context) (bool, error) {
	res, err := c.client.Indices.Exists([]string{c.index})
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	return res.StatusCode == 200, nil
}

// CreateIndex 创建索引
func (c *Client) CreateIndex(ctx context.Context) error {
	exists, err := c.IndexExists(ctx)
	if err != nil {
		return err
	}

	if exists {
		logx.Infof("Index %s already exists", c.index)
		return nil
	}

	res, err := c.client.Indices.Create(
		c.index,
		c.client.Indices.Create.WithBody(strings.NewReader(FileIndexMapping)),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("create index error: %s", res.String())
	}

	logx.Infof("Index %s created successfully", c.index)
	return nil
}

// DeleteIndex 删除索引（谨慎使用）
func (c *Client) DeleteIndex(ctx context.Context) error {
	res, err := c.client.Indices.Delete([]string{c.index})
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("delete index error: %s", res.String())
	}

	logx.Infof("Index %s deleted successfully", c.index)
	return nil
}

// GetIndexStats 获取索引统计信息
func (c *Client) GetIndexStats(ctx context.Context) (map[string]interface{}, error) {
	res, err := c.client.Indices.Stats(c.client.Indices.Stats.WithIndex(c.index))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("get index stats error: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// RefreshIndex 刷新索引（使最近的更改可搜索）
func (c *Client) RefreshIndex(ctx context.Context) error {
	res, err := c.client.Indices.Refresh(c.client.Indices.Refresh.WithIndex(c.index))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("refresh index error: %s", res.String())
	}

	return nil
}

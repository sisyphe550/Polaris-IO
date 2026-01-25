package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"polaris-io/backend/app/search/types"
)

// Search 搜索文件
func (c *Client) Search(ctx context.Context, opts *types.SearchOptions) (*types.FileSearchResult, error) {
	if opts.UserID == 0 {
		return nil, fmt.Errorf("user_id is required")
	}

	// 构建查询
	query := c.buildSearchQuery(opts)

	// 分页
	if opts.Page <= 0 {
		opts.Page = 1
	}
	if opts.PageSize <= 0 {
		opts.PageSize = 20
	}
	if opts.PageSize > 100 {
		opts.PageSize = 100
	}

	from := (opts.Page - 1) * opts.PageSize

	// 构建完整请求
	request := map[string]interface{}{
		"query":            query,
		"from":             from,
		"size":             opts.PageSize,
		"track_total_hits": true,
	}

	// 排序
	if opts.SortBy != "" {
		sortField := opts.SortBy
		if sortField == "name" {
			sortField = "name.keyword" // 使用 keyword 子字段排序
		}
		sortOrder := "asc"
		if opts.SortDesc {
			sortOrder = "desc"
		}
		request["sort"] = []map[string]interface{}{
			{sortField: map[string]string{"order": sortOrder}},
		}
	} else {
		// 默认按相关性排序，然后按更新时间降序
		request["sort"] = []map[string]interface{}{
			{"_score": map[string]string{"order": "desc"}},
			{"update_time": map[string]string{"order": "desc"}},
		}
	}

	// 高亮
	request["highlight"] = map[string]interface{}{
		"fields": map[string]interface{}{
			"name": map[string]interface{}{
				"pre_tags":  []string{"<em>"},
				"post_tags": []string{"</em>"},
			},
		},
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	res, err := c.client.Search(
		c.client.Search.WithContext(ctx),
		c.client.Search.WithIndex(c.index),
		c.client.Search.WithBody(bytes.NewReader(data)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.String())
	}

	// 解析响应
	var result searchResponse
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	// 转换结果
	files := make([]*types.FileDocument, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		files = append(files, &hit.Source)
	}

	return &types.FileSearchResult{
		Total: result.Hits.Total.Value,
		List:  files,
	}, nil
}

// buildSearchQuery 构建搜索查询
func (c *Client) buildSearchQuery(opts *types.SearchOptions) map[string]interface{} {
	// 必须条件：用户ID
	mustClauses := []map[string]interface{}{
		{
			"term": map[string]interface{}{
				"user_id": opts.UserID,
			},
		},
	}

	// 关键词搜索
	if opts.Keyword != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{
						"match": map[string]interface{}{
							"name": map[string]interface{}{
								"query":     opts.Keyword,
								"boost":     2.0, // 文件名匹配权重更高
								"fuzziness": "AUTO",
							},
						},
					},
					{
						"match": map[string]interface{}{
							"name_pinyin": map[string]interface{}{
								"query":     opts.Keyword,
								"fuzziness": "AUTO",
							},
						},
					},
					{
						"wildcard": map[string]interface{}{
							"name.keyword": map[string]interface{}{
								"value":            "*" + opts.Keyword + "*",
								"case_insensitive": true,
							},
						},
					},
				},
				"minimum_should_match": 1,
			},
		})
	}

	// 过滤条件
	filterClauses := []map[string]interface{}{}

	// 扩展名过滤
	if len(opts.Ext) > 0 {
		filterClauses = append(filterClauses, map[string]interface{}{
			"terms": map[string]interface{}{
				"ext": opts.Ext,
			},
		})
	}

	// 是否文件夹过滤
	if opts.IsDir != nil {
		filterClauses = append(filterClauses, map[string]interface{}{
			"term": map[string]interface{}{
				"is_dir": *opts.IsDir,
			},
		})
	}

	// 文件大小过滤
	if opts.MinSize != nil || opts.MaxSize != nil {
		rangeQuery := map[string]interface{}{}
		if opts.MinSize != nil {
			rangeQuery["gte"] = *opts.MinSize
		}
		if opts.MaxSize != nil {
			rangeQuery["lte"] = *opts.MaxSize
		}
		filterClauses = append(filterClauses, map[string]interface{}{
			"range": map[string]interface{}{
				"size": rangeQuery,
			},
		})
	}

	// 构建最终查询
	query := map[string]interface{}{
		"bool": map[string]interface{}{
			"must": mustClauses,
		},
	}

	if len(filterClauses) > 0 {
		query["bool"].(map[string]interface{})["filter"] = filterClauses
	}

	return query
}

// SearchByName 按文件名搜索（简化版）
func (c *Client) SearchByName(ctx context.Context, userID int64, keyword string, page, pageSize int) (*types.FileSearchResult, error) {
	return c.Search(ctx, &types.SearchOptions{
		UserID:   userID,
		Keyword:  keyword,
		Page:     page,
		PageSize: pageSize,
	})
}

// SearchByExt 按扩展名搜索
func (c *Client) SearchByExt(ctx context.Context, userID int64, exts []string, page, pageSize int) (*types.FileSearchResult, error) {
	return c.Search(ctx, &types.SearchOptions{
		UserID:   userID,
		Ext:      exts,
		Page:     page,
		PageSize: pageSize,
	})
}

// CountByUserID 统计用户文件数量
func (c *Client) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"user_id": userID,
			},
		},
	}

	data, err := json.Marshal(query)
	if err != nil {
		return 0, err
	}

	res, err := c.client.Count(
		c.client.Count.WithContext(ctx),
		c.client.Count.WithIndex(c.index),
		c.client.Count.WithBody(bytes.NewReader(data)),
	)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return 0, fmt.Errorf("count error: %s", res.String())
	}

	var result struct {
		Count int64 `json:"count"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return 0, err
	}

	return result.Count, nil
}

// searchResponse ES 搜索响应结构
type searchResponse struct {
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source    types.FileDocument `json:"_source"`
			Highlight map[string][]string `json:"highlight,omitempty"`
		} `json:"hits"`
	} `json:"hits"`
}

package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"polaris-io/backend/app/search/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// IndexDocument 索引文档（创建或更新）
func (c *Client) IndexDocument(ctx context.Context, doc *types.FileDocument) error {
	if doc.ID == "" {
		return fmt.Errorf("document ID is required")
	}

	// 设置更新时间
	doc.UpdateTime = time.Now()

	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	res, err := c.client.Index(
		c.index,
		bytes.NewReader(data),
		c.client.Index.WithDocumentID(doc.ID),
		c.client.Index.WithRefresh("false"), // 不立即刷新，提高性能
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("index document error: %s", res.String())
	}

	logx.Debugf("Document indexed: id=%s, name=%s", doc.ID, doc.Name)
	return nil
}

// UpdateDocument 更新文档（部分更新）
func (c *Client) UpdateDocument(ctx context.Context, id string, fields map[string]interface{}) error {
	if id == "" {
		return fmt.Errorf("document ID is required")
	}

	// 添加更新时间
	fields["update_time"] = time.Now()

	doc := map[string]interface{}{
		"doc": fields,
	}

	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	res, err := c.client.Update(
		c.index,
		id,
		bytes.NewReader(data),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		// 404 表示文档不存在，忽略
		if res.StatusCode == 404 {
			logx.Debugf("Document not found for update: id=%s", id)
			return nil
		}
		return fmt.Errorf("update document error: %s", res.String())
	}

	logx.Debugf("Document updated: id=%s", id)
	return nil
}

// DeleteDocument 删除文档
func (c *Client) DeleteDocument(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("document ID is required")
	}

	res, err := c.client.Delete(
		c.index,
		id,
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		// 404 表示文档不存在，忽略
		if res.StatusCode == 404 {
			logx.Debugf("Document not found for delete: id=%s", id)
			return nil
		}
		return fmt.Errorf("delete document error: %s", res.String())
	}

	logx.Debugf("Document deleted: id=%s", id)
	return nil
}

// GetDocument 获取文档
func (c *Client) GetDocument(ctx context.Context, id string) (*types.FileDocument, error) {
	if id == "" {
		return nil, fmt.Errorf("document ID is required")
	}

	res, err := c.client.Get(
		c.index,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return nil, nil // 文档不存在
		}
		return nil, fmt.Errorf("get document error: %s", res.String())
	}

	var result struct {
		Source types.FileDocument `json:"_source"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Source, nil
}

// BulkIndex 批量索引文档
func (c *Client) BulkIndex(ctx context.Context, docs []*types.FileDocument) error {
	if len(docs) == 0 {
		return nil
	}

	var buf bytes.Buffer

	for _, doc := range docs {
		doc.UpdateTime = time.Now()

		// 操作元数据行
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": c.index,
				"_id":    doc.ID,
			},
		}
		metaData, err := json.Marshal(meta)
		if err != nil {
			return err
		}
		buf.Write(metaData)
		buf.WriteByte('\n')

		// 文档数据行
		docData, err := json.Marshal(doc)
		if err != nil {
			return err
		}
		buf.Write(docData)
		buf.WriteByte('\n')
	}

	res, err := c.client.Bulk(
		bytes.NewReader(buf.Bytes()),
		c.client.Bulk.WithIndex(c.index),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk index error: %s", res.String())
	}

	logx.Debugf("Bulk indexed %d documents", len(docs))
	return nil
}

// BulkDelete 批量删除文档
func (c *Client) BulkDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	var buf bytes.Buffer

	for _, id := range ids {
		meta := map[string]interface{}{
			"delete": map[string]interface{}{
				"_index": c.index,
				"_id":    id,
			},
		}
		metaData, err := json.Marshal(meta)
		if err != nil {
			return err
		}
		buf.Write(metaData)
		buf.WriteByte('\n')
	}

	res, err := c.client.Bulk(
		bytes.NewReader(buf.Bytes()),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk delete error: %s", res.String())
	}

	logx.Debugf("Bulk deleted %d documents", len(ids))
	return nil
}

// DeleteByUserID 删除用户的所有文档
func (c *Client) DeleteByUserID(ctx context.Context, userID int64) error {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"user_id": userID,
			},
		},
	}

	data, err := json.Marshal(query)
	if err != nil {
		return err
	}

	res, err := c.client.DeleteByQuery(
		[]string{c.index},
		bytes.NewReader(data),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("delete by user_id error: %s", res.String())
	}

	logx.Infof("Deleted all documents for user_id=%d", userID)
	return nil
}

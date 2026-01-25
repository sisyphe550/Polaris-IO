package handler

import (
	"context"
	"encoding/json"
	"time"

	"polaris-io/backend/app/search/cmd/job/internal/svc"
	"polaris-io/backend/app/search/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// FileEventHandler 文件事件处理器
type FileEventHandler struct {
	svcCtx *svc.ServiceContext
}

// NewFileEventHandler 创建文件事件处理器
func NewFileEventHandler(svcCtx *svc.ServiceContext) *FileEventHandler {
	return &FileEventHandler{
		svcCtx: svcCtx,
	}
}

// Handle 处理 Kafka 消息
func (h *FileEventHandler) Handle(ctx context.Context, key, value []byte) error {
	var event types.FileEvent
	if err := json.Unmarshal(value, &event); err != nil {
		logx.Errorf("Failed to unmarshal file event: %v", err)
		return err
	}

	logx.Infof("Received file event: type=%s, identity=%s, name=%s", event.EventType, event.Identity, event.Name)

	switch event.EventType {
	case types.EventTypeFileUploaded:
		return h.handleFileUploaded(ctx, &event)
	case types.EventTypeFileUpdated:
		return h.handleFileUpdated(ctx, &event)
	case types.EventTypeFileDeleted:
		return h.handleFileDeleted(ctx, &event)
	default:
		logx.Slowf("Unknown event type: %s", event.EventType)
		return nil
	}
}

// handleFileUploaded 处理文件上传事件
func (h *FileEventHandler) handleFileUploaded(ctx context.Context, event *types.FileEvent) error {
	// 判断是否为文件夹
	isDir := event.Ext == "" && event.Hash == ""

	doc := &types.FileDocument{
		ID:         event.Identity,
		UserID:     event.UserID,
		FileID:     event.FileID,
		Name:       event.Name,
		NamePinyin: "", // TODO: 可以添加拼音转换
		Ext:        event.Ext,
		Size:       event.Size,
		Hash:       event.Hash,
		ParentID:   event.ParentID,
		IsDir:      isDir,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	if err := h.svcCtx.ESClient.IndexDocument(ctx, doc); err != nil {
		logx.Errorf("Failed to index document: %v", err)
		return err
	}

	logx.Infof("Indexed file: identity=%s, name=%s", event.Identity, event.Name)
	return nil
}

// handleFileUpdated 处理文件更新事件（移动/重命名）
func (h *FileEventHandler) handleFileUpdated(ctx context.Context, event *types.FileEvent) error {
	// file_updated 事件总是包含 name 和 parent_id（即使值为空或 0）
	fields := map[string]interface{}{
		"name":      event.Name,
		"parent_id": event.ParentID, // 0 表示根目录
		// TODO: 更新拼音 fields["name_pinyin"] = ...
	}

	if len(fields) == 0 {
		return nil
	}

	if err := h.svcCtx.ESClient.UpdateDocument(ctx, event.Identity, fields); err != nil {
		logx.Errorf("Failed to update document: %v", err)
		return err
	}

	logx.Infof("Updated file: identity=%s", event.Identity)
	return nil
}

// handleFileDeleted 处理文件删除事件
func (h *FileEventHandler) handleFileDeleted(ctx context.Context, event *types.FileEvent) error {
	if err := h.svcCtx.ESClient.DeleteDocument(ctx, event.Identity); err != nil {
		logx.Errorf("Failed to delete document: %v", err)
		return err
	}

	logx.Infof("Deleted file: identity=%s", event.Identity)
	return nil
}

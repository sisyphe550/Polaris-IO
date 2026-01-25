package types

import "time"

// FileDocument ES 文件索引文档结构
type FileDocument struct {
	ID         string    `json:"id"`          // 文件 identity (ES _id)
	UserID     int64     `json:"user_id"`     // 用户ID
	FileID     int64     `json:"file_id"`     // MySQL file_id
	Name       string    `json:"name"`        // 文件名（用于搜索）
	NamePinyin string    `json:"name_pinyin"` // 文件名拼音（用于拼音搜索）
	Ext        string    `json:"ext"`         // 扩展名
	Size       uint64    `json:"size"`        // 文件大小
	Hash       string    `json:"hash"`        // 文件哈希
	ParentID   int64     `json:"parent_id"`   // 父目录ID
	IsDir      bool      `json:"is_dir"`      // 是否为文件夹
	CreateTime time.Time `json:"create_time"` // 创建时间
	UpdateTime time.Time `json:"update_time"` // 更新时间
}

// FileSearchResult 搜索结果
type FileSearchResult struct {
	Total int64           `json:"total"` // 总数
	List  []*FileDocument `json:"list"`  // 文件列表
}

// SearchOptions 搜索选项
type SearchOptions struct {
	UserID   int64    // 用户ID（必须）
	Keyword  string   // 搜索关键词
	Ext      []string // 文件扩展名过滤
	IsDir    *bool    // 是否只搜索文件夹
	MinSize  *uint64  // 最小文件大小
	MaxSize  *uint64  // 最大文件大小
	Page     int      // 页码
	PageSize int      // 每页数量
	SortBy   string   // 排序字段: name, size, create_time, update_time
	SortDesc bool     // 是否降序
}

// 事件类型常量
const (
	EventTypeFileUploaded = "file_uploaded"
	EventTypeFileDeleted  = "file_deleted"
	EventTypeFileUpdated  = "file_updated"
)

// FileEvent Kafka 文件事件（与 pkg/kafka 保持一致）
type FileEvent struct {
	EventType string `json:"event_type"`
	UserID    int64  `json:"user_id"`
	FileID    int64  `json:"file_id"`
	Identity  string `json:"identity"`
	Name      string `json:"name"`
	Hash      string `json:"hash,omitempty"`
	Size      uint64 `json:"size,omitempty"`
	Ext       string `json:"ext,omitempty"`
	ParentID  int64  `json:"parent_id"` // 不使用 omitempty，0 表示根目录
	Timestamp string `json:"timestamp"`
}

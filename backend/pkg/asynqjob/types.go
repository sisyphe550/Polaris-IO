package asynqjob

// 任务类型常量
const (
	// TypeS3Cleanup S3 文件清理任务
	// 当 file_meta 的 ref_count 降为 0 时触发
	TypeS3Cleanup = "s3:cleanup"

	// TypeTrashClear 回收站批量清理任务
	// 大量文件需要彻底删除时触发
	TypeTrashClear = "trash:clear"

	// TypeUploadTimeout 上传超时检测任务
	// 用户开始上传后，定时检测是否完成，未完成则退还配额
	TypeUploadTimeout = "upload:timeout"

	// TypeQuotaRefund 配额退还任务
	// 上传失败或取消时退还用户配额
	TypeQuotaRefund = "quota:refund"
)

// S3CleanupPayload S3 清理任务的 payload
type S3CleanupPayload struct {
	Hash   string `json:"hash"`   // 文件 hash
	S3Key  string `json:"s3Key"`  // S3 对象 key
	UserId int64  `json:"userId"` // 用户 ID（用于日志追踪）
}

// TrashClearPayload 回收站清理任务的 payload
type TrashClearPayload struct {
	UserId    int64 `json:"userId"`    // 用户 ID
	BatchSize int   `json:"batchSize"` // 每批处理数量
}

// UploadTimeoutPayload 上传超时检测任务的 payload
type UploadTimeoutPayload struct {
	UserId    int64  `json:"userId"`    // 用户 ID
	UploadKey string `json:"uploadKey"` // 上传 key
	Hash      string `json:"hash"`      // 文件 hash
	Size      uint64 `json:"size"`      // 文件大小（用于退还配额）
}

// QuotaRefundPayload 配额退还任务的 payload
type QuotaRefundPayload struct {
	UserId int64  `json:"userId"` // 用户 ID
	Size   uint64 `json:"size"`   // 退还大小
	Reason string `json:"reason"` // 退还原因
}

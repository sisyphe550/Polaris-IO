package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FileMeta 文件元数据 (存储在 MongoDB 中，用于秒传)
type FileMeta struct {
	ID         primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Hash       string                 `bson:"hash" json:"hash"`               // 文件 SHA256，唯一索引
	Size       uint64                 `bson:"size" json:"size"`               // 文件大小（字节）
	S3Key      string                 `bson:"s3_key" json:"s3Key"`            // S3 存储路径
	Ext        string                 `bson:"ext" json:"ext"`                 // 文件扩展名
	MimeType   string                 `bson:"mime_type" json:"mimeType"`      // MIME 类型
	RefCount   int64                  `bson:"ref_count" json:"refCount"`      // 引用计数
	ExtAttr    map[string]interface{} `bson:"ext_attr,omitempty" json:"extAttr"` // 扩展属性（图片宽高、视频时长等）
	CreateTime time.Time              `bson:"create_time" json:"createTime"`
	UpdateTime time.Time              `bson:"update_time" json:"updateTime"`
}

// CollectionName 返回集合名称
func (FileMeta) CollectionName() string {
	return "file_meta"
}

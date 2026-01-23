package mongo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrNotFound      = errors.New("file meta not found")
	ErrAlreadyExists = errors.New("file meta already exists")
	ErrRefCountZero  = errors.New("ref count already zero or negative")
)

type FileMetaModel interface {
	// Insert 插入文件元数据
	Insert(ctx context.Context, data *FileMeta) error
	// FindByHash 根据 Hash 查询文件元数据
	FindByHash(ctx context.Context, hash string) (*FileMeta, error)
	// FindById 根据 ID 查询
	FindById(ctx context.Context, id string) (*FileMeta, error)
	// IncrRefCount 增加引用计数
	IncrRefCount(ctx context.Context, hash string, delta int64) error
	// DecrRefCount 减少引用计数
	DecrRefCount(ctx context.Context, hash string, delta int64) error
	// DecrRefCountAndGet 减少引用计数并返回更新后的记录（用于判断是否需要清理 S3）
	DecrRefCountAndGet(ctx context.Context, hash string, delta int64) (*FileMeta, error)
	// DeleteByHash 删除文件元数据 (当引用计数为0时)
	DeleteByHash(ctx context.Context, hash string) error
}

type defaultFileMetaModel struct {
	conn       *mongo.Database
	collection *mongo.Collection
}

// NewFileMetaModel 创建 FileMetaModel 实例
func NewFileMetaModel(db *mongo.Database) FileMetaModel {
	return &defaultFileMetaModel{
		conn:       db,
		collection: db.Collection(FileMeta{}.CollectionName()),
	}
}

// EnsureIndexes 确保索引存在（应用启动时调用）
func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection(FileMeta{}.CollectionName())

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "hash", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("idx_hash"),
		},
		{
			Keys:    bson.D{{Key: "create_time", Value: -1}},
			Options: options.Index().SetName("idx_create_time"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func (m *defaultFileMetaModel) Insert(ctx context.Context, data *FileMeta) error {
	if data.ID.IsZero() {
		data.ID = primitive.NewObjectID()
	}
	now := time.Now()
	data.CreateTime = now
	data.UpdateTime = now
	if data.RefCount == 0 {
		data.RefCount = 1
	}

	_, err := m.collection.InsertOne(ctx, data)
	if mongo.IsDuplicateKeyError(err) {
		return ErrAlreadyExists
	}
	return err
}

func (m *defaultFileMetaModel) FindByHash(ctx context.Context, hash string) (*FileMeta, error) {
	var result FileMeta
	err := m.collection.FindOne(ctx, bson.M{"hash": hash}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &result, nil
}

func (m *defaultFileMetaModel) FindById(ctx context.Context, id string) (*FileMeta, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var result FileMeta
	err = m.collection.FindOne(ctx, bson.M{"_id": objectId}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &result, nil
}

func (m *defaultFileMetaModel) IncrRefCount(ctx context.Context, hash string, delta int64) error {
	filter := bson.M{"hash": hash}
	update := bson.M{
		"$inc": bson.M{"ref_count": delta},
		"$set": bson.M{"update_time": time.Now()},
	}

	result, err := m.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// DecrRefCount 减少引用计数（带保护，防止扣减到负数）
func (m *defaultFileMetaModel) DecrRefCount(ctx context.Context, hash string, delta int64) error {
	// 只有当 ref_count > 0 时才扣减，防止扣减到负数
	filter := bson.M{
		"hash":      hash,
		"ref_count": bson.M{"$gt": 0}, // 保护条件：ref_count > 0
	}
	update := bson.M{
		"$inc": bson.M{"ref_count": -delta},
		"$set": bson.M{"update_time": time.Now()},
	}

	result, err := m.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		// 可能是记录不存在，也可能是 ref_count 已经 <= 0
		// 先检查记录是否存在
		_, findErr := m.FindByHash(ctx, hash)
		if findErr != nil {
			return ErrNotFound
		}
		// 记录存在但 ref_count <= 0
		return ErrRefCountZero
	}
	return nil
}

// DecrRefCountAndGet 减少引用计数并返回更新后的记录
// 使用 FindOneAndUpdate 保证原子性
// 增加保护机制：只有当 ref_count > 0 时才扣减，防止扣减到负数
func (m *defaultFileMetaModel) DecrRefCountAndGet(ctx context.Context, hash string, delta int64) (*FileMeta, error) {
	// 保护条件：只有当 ref_count > 0 时才扣减
	filter := bson.M{
		"hash":      hash,
		"ref_count": bson.M{"$gt": 0},
	}
	update := bson.M{
		"$inc": bson.M{"ref_count": -delta},
		"$set": bson.M{"update_time": time.Now()},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var result FileMeta
	err := m.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// 可能是记录不存在，也可能是 ref_count 已经 <= 0
			// 先检查记录是否存在
			existingMeta, findErr := m.FindByHash(ctx, hash)
			if findErr != nil {
				return nil, ErrNotFound
			}
			// 记录存在但 ref_count <= 0，返回当前记录（不扣减）
			return existingMeta, ErrRefCountZero
		}
		return nil, err
	}
	return &result, nil
}

func (m *defaultFileMetaModel) DeleteByHash(ctx context.Context, hash string) error {
	result, err := m.collection.DeleteOne(ctx, bson.M{"hash": hash})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

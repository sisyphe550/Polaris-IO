// MongoDB 初始化脚本
// 使用方式: mongosh < init.js

// 切换到 polaris_file 数据库
db = db.getSiblingDB('polaris_file');

// 创建 file_meta 集合（文件元数据池）
db.createCollection('file_meta');

// 创建索引
db.file_meta.createIndex({ "hash": 1 }, { unique: true, name: "idx_hash" });
db.file_meta.createIndex({ "create_time": -1 }, { name: "idx_create_time" });

// file_meta 集合结构说明:
// {
//   "_id": ObjectId,
//   "hash": "sha256字符串",          // 文件 SHA256，唯一索引
//   "size": NumberLong,              // 文件大小（字节）
//   "s3_key": "bucket/path/to/file", // S3 存储路径
//   "ext": "pdf",                    // 文件扩展名
//   "mime_type": "application/pdf",  // MIME 类型
//   "ref_count": NumberInt,          // 引用计数（多少用户引用了此文件）
//   "ext_attr": {                    // 扩展属性（可选）
//     "width": 1920,                 // 图片宽度
//     "height": 1080,                // 图片高度
//     "duration": 120                // 视频时长（秒）
//   },
//   "create_time": ISODate,
//   "update_time": ISODate
// }

print("polaris_file database initialized successfully!");

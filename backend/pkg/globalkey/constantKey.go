package globalkey

// =====================================
// 1. 系统级常量 (System Constants)
// =====================================

// CtxJwtUserId 是我们放入 Context 中的 Key
// JWT 中间件解析 Token 后，会把 UserId 放在这个 Key 下
// Logic 层通过 l.ctx.Value(CtxJwtUserId) 获取当前登录用户的 ID
const CtxJwtUserId = "jwtUserId"

// =====================================
// 2. 业务通用常量 (Business Constants)
// =====================================

// 分页默认值
const (
	DefaultPageSize = 10
	MaxPageSize     = 50
)

// 软删除状态 (使用 const 而不是 var，更安全)
const (
	DelStateNo  int64 = 0 // 未删除
	DelStateYes int64 = 1 // 已删除
)

// =====================================
// 3. 白板项目特定常量 (Whiteboard Domain)
// =====================================

// 房间类型 (对应数据库 tinyint)
const (
	RoomTypePublic  int64 = 0 // 公开房间
	RoomTypePrivate int64 = 1 // 加密房间
)

// 房间状态 (对应数据库 tinyint)
const (
	RoomStatusClosed int64 = 0 // 关闭/归档
	RoomStatusActive int64 = 1 // 活跃
)

// 消息类型 (聊天室)
const (
	MsgTypeText  int64 = 0 // 文本消息
	MsgTypeImage int64 = 1 // 图片消息
)

// =====================================
// 4. 时间格式化模板 (Time Format Templates)
// 注意：Go 语言使用固定时间 "2006-01-02 15:04:05" 作为占位符
// =====================================

// DateTimeFormatTplStandardDateTime YYYY-MM-DD HH:MM:SS
const DateTimeFormatTplStandardDateTime = "2006-01-02 15:04:05"

// DateTimeFormatTplStandardDate YYYY-MM-DD
const DateTimeFormatTplStandardDate = "2006-01-02"

// DateTimeFormatTplStandardTime HH:MM:SS
const DateTimeFormatTplStandardTime = "15:04:05"

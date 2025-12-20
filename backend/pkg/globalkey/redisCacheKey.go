package globalkey

/**
Redis Key 定义规范：
1. 使用项目名为前缀 (whiteboard) 防止冲突
2. 使用冒号分隔层级
3. 结尾明确 Key 的参数类型 (%d, %s)
*/

const (
	// ============================
	// 用户相关 (User)
	// ============================

	// CacheUserTokenKey 用户登陆的 token
	// Key: whiteboard:user:token:uid
	// Value: token string
	CacheUserTokenKey = "whiteboard:user:token:%d"

	// ============================
	// 房间与白板相关 (Room & Whiteboard)
	// ============================

	// CacheRoomInfoKey 房间基础信息缓存 (用于高频查询房间是否存在/加密)
	// Key: whiteboard:room:info:uuid
	// Value: json string
	CacheRoomInfoKey = "whiteboard:room:info:%s"

	// ============================
	// WebSocket 协同相关 (核心)
	// ============================

	// RedisRoomChannelPrefix Pub/Sub 广播通道前缀
	// 完整 Key: whiteboard:channel:room:{roomId}
	RedisRoomChannelPrefix = "whiteboard:channel:room:"

	// RedisRoomOnlinePrefix 房间在线用户列表 (Set 集合)
	// Key: whiteboard:room:users:{roomId}
	// Value: userId
	RedisRoomOnlinePrefix = "whiteboard:room:users:"

	// RedisRoomHistoryPrefix 白板绘图指令历史队列 (List)
	// Key: whiteboard:room:history:{roomId}
	// Value: draw_command_json
	RedisRoomHistoryPrefix = "whiteboard:room:history:"
)

package xerr

// 成功返回
const OK uint32 = 200

/**(前3位代表业务,后三位代表具体功能)**/

// 全局错误码
const SERVER_COMMON_ERROR uint32 = 100001           // 服务器开小差了
const REUQEST_PARAM_ERROR uint32 = 100002           // 参数错误
const TOKEN_EXPIRE_ERROR uint32 = 100003            // Token过期
const TOKEN_GENERATE_ERROR uint32 = 100004          // Token生成错误
const DB_ERROR uint32 = 100005                      // 数据库错误
const DB_UPDATE_AFFECTED_ZERO_ERROR uint32 = 100006 // 数据库更新影响行数为0

// 用户模块 (200xxx)
const USER_NOT_EXIST uint32 = 200001       // 用户不存在
const USER_ALREADY_EXISTS uint32 = 200002  // 用户已存在
const USER_PASSWORD_ERROR uint32 = 200003  // 密码错误
const USER_QUOTA_EXCEEDED uint32 = 200004  // 存储空间不足
const USER_QUOTA_NOT_EXIST uint32 = 200005 // 配额记录不存在

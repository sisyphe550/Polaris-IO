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

// 文件模块 (300xxx)
const FILE_NOT_EXIST uint32 = 300001           // 文件不存在
const FILE_ALREADY_EXISTS uint32 = 300002      // 文件已存在
const FILE_UPLOAD_FAILED uint32 = 300003       // 文件上传失败
const FILE_DOWNLOAD_FAILED uint32 = 300004     // 文件下载失败
const FILE_DELETE_FAILED uint32 = 300005       // 文件删除失败
const FILE_MOVE_FAILED uint32 = 300006         // 文件移动失败
const FILE_COPY_FAILED uint32 = 300007         // 文件复制失败
const FILE_RENAME_FAILED uint32 = 300008       // 文件重命名失败
const FOLDER_NOT_EXIST uint32 = 300009         // 文件夹不存在
const FOLDER_ALREADY_EXISTS uint32 = 300010    // 文件夹已存在
const FOLDER_CREATE_FAILED uint32 = 300011     // 文件夹创建失败
const FILE_NAME_INVALID uint32 = 300012        // 文件名无效
const FILE_NAME_DUPLICATE uint32 = 300013      // 同目录下文件名重复
const FILE_PARENT_NOT_EXIST uint32 = 300014    // 父目录不存在
const FILE_CANNOT_MOVE_TO_SELF uint32 = 300015 // 不能移动到自身或子目录
const FILE_IN_TRASH uint32 = 300016            // 文件在回收站中
const FILE_NOT_IN_TRASH uint32 = 300017        // 文件不在回收站中
const FILE_RESTORE_FAILED uint32 = 300018      // 文件恢复失败
const FILE_META_NOT_FOUND uint32 = 300019      // 文件元数据不存在
const S3_PRESIGN_FAILED uint32 = 300020        // S3 预签名失败
const S3_UPLOAD_FAILED uint32 = 300021         // S3 上传失败
const S3_DELETE_FAILED uint32 = 300022         // S3 删除失败

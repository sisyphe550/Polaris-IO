package xerr

var message map[uint32]string

func init() {
	message = make(map[uint32]string)
	message[OK] = "SUCCESS"
	message[SERVER_COMMON_ERROR] = "服务器开小差啦,稍后再来试一试"
	message[REUQEST_PARAM_ERROR] = "参数错误"
	message[TOKEN_EXPIRE_ERROR] = "token失效，请重新登陆"
	message[TOKEN_GENERATE_ERROR] = "生成token失败"
	message[DB_ERROR] = "数据库繁忙,请稍后再试"
	message[DB_UPDATE_AFFECTED_ZERO_ERROR] = "更新数据影响行数为0"

	// 用户模块
	message[USER_NOT_EXIST] = "用户不存在"
	message[USER_ALREADY_EXISTS] = "该手机号已注册"
	message[USER_PASSWORD_ERROR] = "密码错误"
	message[USER_QUOTA_EXCEEDED] = "存储空间不足"
	message[USER_QUOTA_NOT_EXIST] = "配额记录不存在"

	// 文件模块
	message[FILE_NOT_EXIST] = "文件不存在"
	message[FILE_ALREADY_EXISTS] = "文件已存在"
	message[FILE_UPLOAD_FAILED] = "文件上传失败"
	message[FILE_DOWNLOAD_FAILED] = "文件下载失败"
	message[FILE_DELETE_FAILED] = "文件删除失败"
	message[FILE_MOVE_FAILED] = "文件移动失败"
	message[FILE_COPY_FAILED] = "文件复制失败"
	message[FILE_RENAME_FAILED] = "文件重命名失败"
	message[FOLDER_NOT_EXIST] = "文件夹不存在"
	message[FOLDER_ALREADY_EXISTS] = "文件夹已存在"
	message[FOLDER_CREATE_FAILED] = "文件夹创建失败"
	message[FILE_NAME_INVALID] = "文件名无效"
	message[FILE_NAME_DUPLICATE] = "同目录下已存在同名文件"
	message[FILE_PARENT_NOT_EXIST] = "父目录不存在"
	message[FILE_CANNOT_MOVE_TO_SELF] = "不能移动到自身或子目录"
	message[FILE_IN_TRASH] = "文件在回收站中"
	message[FILE_NOT_IN_TRASH] = "文件不在回收站中"
	message[FILE_RESTORE_FAILED] = "文件恢复失败"
	message[FILE_META_NOT_FOUND] = "文件元数据不存在"
	message[S3_PRESIGN_FAILED] = "获取上传/下载链接失败"
	message[S3_UPLOAD_FAILED] = "文件上传到存储服务失败"
	message[S3_DELETE_FAILED] = "从存储服务删除文件失败"

	// 分享模块
	message[SHARE_NOT_EXIST] = "分享不存在"
	message[SHARE_EXPIRED] = "分享已过期"
	message[SHARE_CODE_ERROR] = "提取码错误"
	message[SHARE_CANCELLED] = "分享已取消"
	message[SHARE_BANNED] = "分享已被封禁"
	message[SHARE_FILE_NOT_EXIST] = "分享的文件不存在"
	message[SHARE_CREATE_FAILED] = "创建分享失败"
	message[SHARE_ALREADY_EXISTS] = "该文件已被分享"
	message[SHARE_SAVE_FAILED] = "保存分享失败"
	message[SHARE_PERMISSION_DENIED] = "无权操作此分享"
}

func MapErrMsg(errcode uint32) string {
	if msg, ok := message[errcode]; ok {
		return msg
	} else {
		return "服务器开小差啦,稍后再来试一试"
	}
}

func IsCodeErr(errcode uint32) bool {
	if _, ok := message[errcode]; ok {
		return true
	} else {
		return false
	}
}

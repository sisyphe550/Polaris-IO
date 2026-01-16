/*
 Source Database       : polaris_file
*/

CREATE DATABASE IF NOT EXISTS `polaris_file`;
USE `polaris_file`;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- 1. 用户仓库表 (文件目录树)
-- ----------------------------
DROP TABLE IF EXISTS `user_repository`;
CREATE TABLE `user_repository` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `identity` varchar(36) NOT NULL DEFAULT '' COMMENT '文件唯一标识',
  `hash` varchar(64) NOT NULL DEFAULT '' COMMENT '文件指纹',
  `user_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '用户ID',
  `parent_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '父目录ID',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '文件名',
  `ext` varchar(30) NOT NULL DEFAULT '' COMMENT '扩展名',
  `size` bigint unsigned NOT NULL DEFAULT '0' COMMENT '文件大小',
  `path` varchar(255) NOT NULL DEFAULT '' COMMENT '物理路径',
  
  -- 标准通用字段
  `version` bigint unsigned NOT NULL DEFAULT '0' COMMENT '乐观锁版本号',
  `del_state` tinyint(1) NOT NULL DEFAULT '0' COMMENT '删除状态 0:正常 1:已删除',
  -- `delete_time` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  `delete_time` bigint unsigned NOT NULL DEFAULT '0' COMMENT '删除时间戳(0:未删 >0:已删)',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  PRIMARY KEY (`id`),
  KEY `idx_user_parent` (`user_id`, `parent_id`) USING BTREE,
  UNIQUE KEY `idx_identity` (`identity`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户文件仓库表';

SET FOREIGN_KEY_CHECKS = 1;
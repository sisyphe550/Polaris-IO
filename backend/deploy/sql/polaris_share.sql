/*
 Source Database       : polaris_share
*/

CREATE DATABASE IF NOT EXISTS `polaris_share`;
USE `polaris_share`;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- 1. 分享记录表
-- ----------------------------
DROP TABLE IF EXISTS `share`;
CREATE TABLE `share` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `identity` varchar(36) NOT NULL DEFAULT '' COMMENT '分享唯一标识',
  `user_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '分享者ID',
  `repository_identity` varchar(36) NOT NULL DEFAULT '' COMMENT '关联文件Identity',
  `code` varchar(10) NOT NULL DEFAULT '' COMMENT '提取码',
  `click_num` int unsigned NOT NULL DEFAULT '0' COMMENT '点击次数',
  `expired_time` int unsigned NOT NULL DEFAULT '0' COMMENT '失效时间(0永久)',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '业务状态 0:正常 1:已封禁', 
  
  -- 标准通用字段
  `version` bigint unsigned NOT NULL DEFAULT '0' COMMENT '乐观锁版本号',
  `del_state` tinyint(1) NOT NULL DEFAULT '0' COMMENT '删除状态 0:正常 1:已取消(软删)',
  -- `delete_time` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  `delete_time` bigint unsigned NOT NULL DEFAULT '0' COMMENT '删除时间戳(0:未删 >0:已删)',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  PRIMARY KEY (`id`),
  KEY `idx_identity` (`identity`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='分享记录表';

SET FOREIGN_KEY_CHECKS = 1;
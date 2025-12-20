CREATE DATABASE IF NOT EXISTS whiteboard;
USE whiteboard;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for rooms
-- ----------------------------
DROP TABLE IF EXISTS `rooms`;
CREATE TABLE `rooms` (
  `id` char(36) NOT NULL COMMENT '房间UUID',
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '房间名称',
  `owner_id` bigint(20) UNSIGNED NOT NULL DEFAULT '0' COMMENT '房主ID',
  `type` tinyint(4) NOT NULL DEFAULT '0' COMMENT '房间类型: 0:公开, 1:加密',
  `password` varchar(50) NOT NULL DEFAULT '' COMMENT '房间密码(仅加密房间有效)',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态: 1:活跃, 0:关闭/归档',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `delete_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间',
  `del_state` tinyint(4) NOT NULL DEFAULT '0' COMMENT '删除状态: 0-正常, 1-已删除',
  `version` bigint(20) NOT NULL DEFAULT '0' COMMENT '版本号(乐观锁)',
  PRIMARY KEY (`id`),
  KEY `idx_owner_id` (`owner_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='房间表';

SET FOREIGN_KEY_CHECKS = 1;
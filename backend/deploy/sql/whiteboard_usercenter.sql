CREATE DATABASE IF NOT EXISTS whiteboard;
USE whiteboard;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for users
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `username` varchar(50) NOT NULL DEFAULT '' COMMENT '用户账号',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT '密码Hash',
  `avatar` varchar(255) NOT NULL DEFAULT '' COMMENT '头像(RustFS URL)',
  `info` varchar(255) NOT NULL DEFAULT '' COMMENT '个人简介',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `delete_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间',
  `del_state` tinyint(4) NOT NULL DEFAULT '0' COMMENT '删除状态: 0-正常, 1-已删除',
  `version` bigint(20) NOT NULL DEFAULT '0' COMMENT '版本号(乐观锁)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`, `delete_time`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

SET FOREIGN_KEY_CHECKS = 1;
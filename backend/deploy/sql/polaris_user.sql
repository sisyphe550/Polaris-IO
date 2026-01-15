/*
 Source Database       : polaris_user
*/

CREATE DATABASE IF NOT EXISTS `polaris_user`;
USE `polaris_user`;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- 1. 用户基本信息表
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `mobile` char(11) NOT NULL DEFAULT '' COMMENT '手机号',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '用户昵称',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT '加密后的密码',
  `avatar` varchar(255) NOT NULL DEFAULT '' COMMENT '头像地址',
  `info` varchar(255) NOT NULL DEFAULT '' COMMENT '个人简介',
  
  -- 标准通用字段
  `version` bigint unsigned NOT NULL DEFAULT '0' COMMENT '乐观锁版本号',
  `del_state` tinyint(1) NOT NULL DEFAULT '0' COMMENT '删除状态 0:正常 1:已删除',
  `delete_time` bigint unsigned NOT NULL DEFAULT '0' COMMENT '删除时间戳(0:未删 >0:已删)',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  PRIMARY KEY (`id`),
  -- 【核心修正】联合唯一索引：手机号 + 删除时间
  -- 只有当 delete_time 都是 0 (活着) 时，手机号才不允许重复
  UNIQUE KEY `idx_mobile` (`mobile`, `delete_time`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户表';

-- ----------------------------
-- 2. 用户容量配额表
-- ----------------------------
DROP TABLE IF EXISTS `user_quota`;
CREATE TABLE `user_quota` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '用户ID',
  `total_size` bigint unsigned NOT NULL DEFAULT '0' COMMENT '总容量(字节)',
  `used_size` bigint unsigned NOT NULL DEFAULT '0' COMMENT '已用容量(字节)',
  
  `version` bigint unsigned NOT NULL DEFAULT '0' COMMENT '乐观锁版本号',
  `del_state` tinyint(1) NOT NULL DEFAULT '0' COMMENT '删除状态 0:正常 1:已删除',
  `delete_time` bigint unsigned NOT NULL DEFAULT '0' COMMENT '删除时间戳',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  PRIMARY KEY (`id`),
  -- 这里其实 user_id 本身就是唯一的（因为新用户有新ID），但加上 delete_time 更保险
  UNIQUE KEY `idx_user_id` (`user_id`, `delete_time`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户容量表';

SET FOREIGN_KEY_CHECKS = 1;
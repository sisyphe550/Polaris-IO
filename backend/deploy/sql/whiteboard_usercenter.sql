CREATE DATABASE IF NOT EXISTS whiteboard;
USE whiteboard;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- 1. 用户基本信息表 (User Profile)
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `mobile` char(11) NOT NULL DEFAULT '' COMMENT '手机号(冗余字段,用于展示/业务检索)',
  `nickname` varchar(50) NOT NULL DEFAULT '' COMMENT '用户昵称',
  `avatar` varchar(255) NOT NULL DEFAULT '' COMMENT '头像(RustFS URL)',
  `info` varchar(255) NOT NULL DEFAULT '' COMMENT '个人简介',
  
  -- 公共字段
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `delete_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间',
  `del_state` tinyint(4) NOT NULL DEFAULT '0' COMMENT '删除状态: 0-正常, 1-已删除',
  `version` bigint(20) NOT NULL DEFAULT '0' COMMENT '版本号(乐观锁)',
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_mobile` (`mobile`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户基本信息表';

-- ----------------------------
-- 2. 用户授权表 (User Authentication)
-- ----------------------------
DROP TABLE IF EXISTS `user_auth`;
CREATE TABLE `user_auth` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` bigint(20) UNSIGNED NOT NULL DEFAULT '0' COMMENT '用户ID',
  `auth_type` varchar(20) NOT NULL DEFAULT 'mobile' COMMENT '认证类型: mobile, wxMini, system',
  `auth_key` varchar(64) NOT NULL DEFAULT '' COMMENT '认证唯一标识: 手机号, OpenID, UnionID',
  `credential` varchar(255) NOT NULL DEFAULT '' COMMENT '认证凭证: 密码Hash, 或第三方Token',
  
  -- 公共字段
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `delete_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间',
  `del_state` tinyint(4) NOT NULL DEFAULT '0' COMMENT '删除状态: 0-正常, 1-已删除',
  `version` bigint(20) NOT NULL DEFAULT '0' COMMENT '版本号(乐观锁)',
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_type_key` (`auth_type`, `auth_key`, `del_state`) USING BTREE,
  KEY `idx_user_id` (`user_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户授权表';

SET FOREIGN_KEY_CHECKS = 1;

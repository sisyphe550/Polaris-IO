CREATE DATABASE IF NOT EXISTS whiteboard;
USE whiteboard;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for snapshots (白板快照)
-- ----------------------------
DROP TABLE IF EXISTS `snapshots`;
CREATE TABLE `snapshots` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `room_id` char(36) NOT NULL DEFAULT '' COMMENT '关联房间ID',
  -- 关键修改：将原有的 version 改名为 revision，避免与乐观锁冲突
  `revision` int(11) NOT NULL DEFAULT '1' COMMENT '快照版本号(业务)',
  `data_json` longtext COMMENT 'Fabric.js 全量 JSON 数据',
  `preview_url` varchar(255) NOT NULL DEFAULT '' COMMENT '白板缩略图 (RustFS URL)',
  
  -- Looklook 标准字段
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `delete_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间',
  `del_state` tinyint(4) NOT NULL DEFAULT '0' COMMENT '删除状态: 0-正常, 1-已删除',
  `version` bigint(20) NOT NULL DEFAULT '0' COMMENT '版本号(乐观锁)',
  
  PRIMARY KEY (`id`),
  KEY `idx_room_id` (`room_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='白板快照表';

-- ----------------------------
-- Table structure for chat_messages (聊天记录)
-- ----------------------------
DROP TABLE IF EXISTS `chat_messages`;
CREATE TABLE `chat_messages` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `room_id` char(36) NOT NULL DEFAULT '' COMMENT '房间ID',
  `user_id` bigint(20) UNSIGNED NOT NULL DEFAULT '0' COMMENT '发送者ID',
  `content` text COMMENT '消息内容',
  `msg_type` tinyint(4) NOT NULL DEFAULT '0' COMMENT '消息类型: 0:文本, 1:图片',
  
  -- Looklook 标准字段 (虽然聊天记录通常不删，但为了Model生成不报错，建议加上)
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `delete_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间',
  `del_state` tinyint(4) NOT NULL DEFAULT '0' COMMENT '删除状态: 0-正常, 1-已删除',
  `version` bigint(20) NOT NULL DEFAULT '0' COMMENT '版本号(乐观锁)',
  
  PRIMARY KEY (`id`),
  KEY `idx_room_create` (`room_id`, `create_time`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='聊天记录表';

SET FOREIGN_KEY_CHECKS = 1;
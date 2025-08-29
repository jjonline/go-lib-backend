-- MySQL队列系统表结构
-- 基于Redis队列原理实现，支持普通队列、延迟队列和保留队列

-- 队列任务表（相当于Redis List）
CREATE TABLE `queue_jobs` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `queue_name` varchar(191) NOT NULL COMMENT '队列名称',
    `payload` longtext NOT NULL COMMENT '任务载荷JSON',
    `attempts` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '已尝试次数',
    `reserved_at` int(10) unsigned DEFAULT NULL COMMENT '保留时间戳',
    `available_at` int(10) unsigned NOT NULL COMMENT '可执行时间戳',
    `created_at` int(10) unsigned NOT NULL COMMENT '创建时间戳',
    PRIMARY KEY (`id`),
    KEY `idx_queue_name` (`queue_name`),
    KEY `idx_available_at` (`available_at`),
    KEY `idx_reserved_at` (`reserved_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='队列任务表';

-- 失败任务表（可选，用于存储失败的任务）
CREATE TABLE `queue_failed_jobs` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `queue_name` varchar(191) NOT NULL COMMENT '队列名称',
    `payload` longtext NOT NULL COMMENT '任务载荷JSON',
    `exception` longtext NOT NULL COMMENT '异常信息',
    `failed_at` int(10) unsigned NOT NULL COMMENT '失败时间戳',
    PRIMARY KEY (`id`),
    KEY `idx_queue_name` (`queue_name`),
    KEY `idx_failed_at` (`failed_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='失败任务表';

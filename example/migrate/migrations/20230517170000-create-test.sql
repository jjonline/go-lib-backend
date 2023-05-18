-- +migrate Up
CREATE TABLE IF NOT EXISTS `test`
(
    `id`         int(10) unsigned NOT NULL AUTO_INCREMENT,
    `name`       varchar(128) NOT NULL DEFAULT '' COMMENT '廣告位類型',
    `created_at` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '創建時間',
    `updated_at` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新時間',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB COLLATE utf8mb4_general_ci DEFAULT CHARSET=utf8mb4 COMMENT='test';

-- +migrate Down
drop table if exists `test`;



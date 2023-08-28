CREATE TABLE `event_info`
(
    `id`            int(11) unsigned not null auto_increment comment 'id',
    `status`        varchar(64) not null default '',

    `sport_type`    varchar(64) null default '',
    `created_at`  TIMESTAMP        NULL     DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  TIMESTAMP        NULL     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    `deleted_at`  TIMESTAMP        NULL     ,
    PRIMARY KEY (`id`),
    KEY `idx_open_id` (`open_id`),
    UNIQUE KEY `uniq_user` (`open_id`, `sport_type`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4
  comment = '微信用户表';
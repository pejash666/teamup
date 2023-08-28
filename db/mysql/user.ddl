CREATE TABLE `wechat_user_info`
(
    `id`            int(11) unsigned not null auto_increment comment 'id',
    `open_id`       varchar(64) not null default '' comment '用户openid',
    `union_id`      varchar(64) null default '',
    `session_key`   varchar(255) not null default '',
    `nickname`      varchar(64) null default '',
    `avatar`        varchar(255) null default '',
    `gender`        varchar(64) null default '',
    `phone_number`  varchar(64) null default '',
    `joined_times`  int unsigned null default 0,
    `joined_group`  varchar(64) null default '',
    `preference`    varchar(64) null default '',
    `tags`          varchar(64) null default '',
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
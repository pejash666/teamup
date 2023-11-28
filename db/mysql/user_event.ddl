CREATE TABLE `user_event`
(
    `id`            int(11) unsigned not null auto_increment comment 'id',
    `user_id`       int(11) unsigned not null default 0 comment '用户id',
    `open_id`       varchar(64) not null default '' comment '用户openid',
    `sport_type`    varchar(64) not null default '' comment '运动类型',
    `event_id`      int(11) not null default '' comment '时间表ID',
    `created_at`  TIMESTAMP        NULL     DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  TIMESTAMP        NULL     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    `deleted_at`  TIMESTAMP        NULL     ,
    PRIMARY KEY (`id`),
    KEY `idx_open_id` (`open_id`),
    KEY `idx_event_id` (`event_id`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4
  comment = '用户事件表';
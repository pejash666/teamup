CREATE TABLE `player_level`
(
    `id`            int(11) unsigned not null auto_increment comment 'id',
    `open_id`       varchar(64) not null default '' comment '用户openid',
    `sport_type`    varchar(64) not null default '' comment '运动类型',
    `calibrated`   smallint(1)  not null default 0 comment '是否已定级',
    `level`         smallint     not null default 0 comment '级别',
    `created_at`  TIMESTAMP        NULL     DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  TIMESTAMP        NULL     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    `deleted_at`  TIMESTAMP        NULL     ,
    PRIMARY KEY (`id`),
    KEY `idx_open_id` (`open_id`),
    UNIQUE KEY `uniq_user` (`open_id`, `sport_type`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4
  comment = '用户级别表';
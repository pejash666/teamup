CREATE TABLE `event_info`
(
    `id`            int(11) unsigned not null auto_increment comment '主键',
    `status`        varchar(64) not null default '' comment 'game状态',
    `is_host`       smallint not null default 0 comment '是否为场地创建活动',
    `creator`       varchar(64) not null default '' comment '创建者',
    `sport_type`    varchar(64) not null default '' comment '运动类型',
    `match_type`    varchar(64) not null default '' comment '比赛类型',
    `game_type`     varchar(64) not null default '' comment '对局类型',
    `is_public`     smallint not null default 0 comment '公开比赛',
    `is_booked`     smallint not null default 0 comment '已订场地',
    `lowest_level`  smallint not null default 0 comment '适合的最低级别',
    `highest_level` smallint not null default 0 comment '适合的最高级别',
    `Date`          varchar(64) not null default '' comment '日期，20060102',
    `city`          varchar(64) not null default '' comment '城市',
    `name`          varchar(64) not null default '' comment '标题',
    `desc`          varchar(64) not null default '' comment '描述',
    `start_time`    int(11) not null default 0  comment '开始的时间戳',
    `end_time`      int(11) not null default 0  comment '结束的时间戳',
    `field_name`    varchar(64) not null default '' comment '场地名字',
    `max_player_num`    smallint not null default 0 comment '最多几人',
    `current_player_num` smallint not null default 0 comment '当前几人',
    `current_player` varchar(64) not null default '' comment '参与的用户id',
    `price`         smallint not null default 0 comment '价格',
    `created_at`  TIMESTAMP        NOT NULL     DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  TIMESTAMP        NOT NULL     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    `deleted_at`  TIMESTAMP        NULL     ,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4
  comment = '事件表';
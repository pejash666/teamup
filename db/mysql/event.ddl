CREATE TABLE `event_info`
(
    `id`            int(11) unsigned not null auto_increment comment '主键',
    `status`        varchar(64) not null default '' comment 'game状态',
    `is_host`       smallint not null default 0 comment '是否为场地创建活动',
    `organization_id` int(11) unsigned not null default 0 comment '组织ID',
    `creator`       varchar(64) not null default '' comment '创建者',
    `sport_type`    varchar(64) not null default '' comment '运动类型',
    `match_type`    varchar(64) not null default '' comment '比赛类型',
    `scorers`       varchar(256) not null default '' comment '记分员',
    `game_type`     varchar(64) not null default '' comment '对局类型',
    `score_rule`    varchar(64) not null default '' comment '记分规则',
    `is_public`     smallint not null default 0 comment '公开比赛',
    `is_booked`     smallint not null default 0 comment '已订场地',
    `lowest_level`  smallint not null default 0 comment '适合的最低级别',
    `highest_level` smallint not null default 0 comment '适合的最高级别',
    `Date`          varchar(64) not null default '' comment '日期，20060102',
    `weekday`       varchar(64) not null default '' comment '星期几',
    `city`          varchar(64) not null default '' comment '城市',
    `name`          varchar(64) not null default '' comment '标题',
    `desc`          varchar(64) not null default '' comment '描述',
    `start_time`    int(11) not null default 0  comment '开始的时间戳',
    `start_time_str` varchar(64) not null default '' comment '开始时间，比如15:04',
    `end_time`      int(11) not null default 0  comment '结束的时间戳',
    `end_time_str`  varchar(64) not null default '' comment '结束事件, 比如15:04',
    `field_name`    varchar(64) not null default '' comment '场地名字',
    `longitude`     varchar(255) not null default '' comment '经度',
    `latitude`      varchar(255) not null default '' comment '纬度',
    `field_type`    varchar(64) not null default '' comment '场地类型',
    `max_player_num`    smallint not null default 0 comment '最多几人',
    `current_player_num` smallint not null default 0 comment '当前几人',
    `current_player` varchar(64) not null default '' comment '参与的用户id',
    `price`         smallint not null default 0 comment '价格',
    `event_image`   varchar(255) not null default '' comment '活动图片',
    `created_at`  TIMESTAMP        NOT NULL     DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  TIMESTAMP        NOT NULL     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    `deleted_at`  TIMESTAMP        NULL     ,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4
  comment = '事件表';
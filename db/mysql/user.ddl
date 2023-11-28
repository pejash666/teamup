CREATE TABLE `wechat_user_info`
(
    `id`            int(11) unsigned not null auto_increment comment 'id',
    `open_id`       varchar(64) not null default '' comment '用户openid',
    `sport_type`    varchar(64) null default '' comment '运动类型',
    `is_primary`    smallint not null default 0 comment '是否为主记录',
    `union_id`      varchar(64) null default '' comment 'unionID',
    `session_key`   varchar(255) not null default '' comment 'session_key',
    `nickname`      varchar(64) not null default '' comment '用户昵称',
    `is_host`       smallint not null default 0 comment '是否是主理人',
    `organization_id` int(11) not null default 0 comment '组织ID',
    `avatar`        varchar(255) not null default '' comment '头像',
    `gender`        varchar(64) not null default '' comment '性别',
    `phone_number`  varchar(64) not null default '' comment '电话号码',
    `joined_times`  int unsigned not null default 0 comemnt '参与次数',
    `joined_event`  string not null default '' comment '参与的event，一个string的列表',
    `joined_organization`  varchar(64) null default '' comment '参与的组织',
    `preference`    varchar(64) null default '' comment '偏好',
    `tags`          varchar(64) null default '' comment '标签',
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
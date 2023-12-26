CREATE TABLE `organization_info`
(
    `id`            int(11) unsigned not null auto_increment comment 'id',
    `name`          varchar(64) not null default '' comment '组织名',
    `city`          varchar(64) not null default '' comment '城市',
    `address`       varchar(64) not null default '' comment '地址',
    `open_id`  varchar(64) not null default '' comment '主理人open_id',
    `contact`       varchar(64) not null default '' comment '联系方式',
    `logo`          varchar(64) not null default '' comment '图标链接',
    `sport_type`    varchar(64) not null default '' comment '运动类型',
    `is_approved`   smallint not null default 0 comment '是否审批通过',
    `created_at`  TIMESTAMP        NOT NULL     DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  TIMESTAMP        NOT NULL     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    `deleted_at`  TIMESTAMP        NULL     ,
    PRIMARY KEY (`id`),
    KEY `idx_open_id` (`open_id`),
    UNIQUE KEY `uniq_user` (`open_id`, `sport_type`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4
  comment = '组织信息表';
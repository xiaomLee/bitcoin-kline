use dev;

DROP TABLE IF EXISTS `kline`;
CREATE TABLE `kline` (
     `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
     `coinType` varchar(10) NOT NULL DEFAULT '' COMMENT '币种',
     `high` varchar(20)  NOT NULL COMMENT '最高报价',
     `low` varchar(20) NOT NULL DEFAULT '0' COMMENT '最低报价',
     `open` varchar(20) NOT NULL COMMENT '开盘价',
     `close` varchar(20) NOT NULL COMMENT '收盘价',
     `createTime` bigint NOT NULL COMMENT '时间',
     `updateTime` bigint NOT NULL COMMENT '最后更新时间',
     `timeScale` varchar(20) NOT NULL COMMENT '分时图刻度',
     `origin` tinyint NOT NULL COMMENT '是否原始数据：1:是，0：否',
     `originPrice` varchar(20)  NOT NULL DEFAULT '' COMMENT '原始报价',
     `volume` varchar(32)  NOT NULL DEFAULT '0' COMMENT '交易笔数',
     PRIMARY KEY (`id`) USING BTREE,
     UNIQUE KEY `coin_time_scale` (`coinType`,`timeScale`,`createTime`) USING BTREE
) ENGINE=InnoDB COMMENT='kline数据';


DROP TABLE IF EXISTS `tick_cache`;
CREATE TABLE `tick_cache` (
     `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
     `coinType` varchar(10) NOT NULL DEFAULT '' COMMENT '币种',
     `high` varchar(20)  NOT NULL COMMENT '最高报价',
     `low` varchar(20) NOT NULL DEFAULT '0' COMMENT '最低报价',
     `open` varchar(20) NOT NULL COMMENT '开盘价',
     `close` varchar(20) NOT NULL COMMENT '收盘价',
     `createTime` bigint NOT NULL COMMENT '时间',
     `updateTime` bigint NOT NULL COMMENT '最后更新时间',
     `timeScale` varchar(20) NOT NULL COMMENT '分时图刻度',
     `origin` tinyint NOT NULL COMMENT '是否原始数据：1:是，0：否',
     `originPrice` varchar(20)  NOT NULL DEFAULT '' COMMENT '原始报价',
     `volume` varchar(32)  NOT NULL DEFAULT '0' COMMENT '交易笔数',
     PRIMARY KEY (`id`) USING BTREE,
     UNIQUE KEY `coin_time_scale` (`coinType`,`timeScale`,`createTime`) USING BTREE
) ENGINE=InnoDB COMMENT='外部实时报价数据';
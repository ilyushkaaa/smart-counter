SET NAMES utf8;
SET time_zone = '+00:00';
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';

CREATE TABLE IF NOT EXISTS `expressions`(
    `id` varchar(64) NOT NULL,
    `body` varchar(255) NOT NULL,
    `result` int,
    `creation_time` varchar(50) NOT NULL,
    `finish_time` varchar(50),
    `is_finished` boolean NOT NULL,
    `is_successful` boolean,
    PRIMARY KEY (`id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8;


CREATE TABLE IF NOT EXISTS `computing_resource`(
    `id` int NOT NULL AUTO_INCREMENT,
    `name` varchar(255) NOT NULL,
    `status` varchar(20) NOT NULL,
    `parallel_computers_num` int NOT NULL,
    `last_ping_time` varchar(50),
    PRIMARY KEY (`id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8;


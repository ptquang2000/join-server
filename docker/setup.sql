-- show query logs
SET GLOBAL log_output = 'TABLE';
SET GLOBAL general_log = 'ON';
SELECT * FROM `mysql`.`general_log` LIMIT 1000;

-- clear query logs
-- SET GLOBAL general_log= 'OFF';
-- TRUNCATE table mysql.general_log;

CREATE DATABASE lorawan;

-- EMQX authentication
CREATE TABLE `lorawan`.`join_servers` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(100) DEFAULT NULL,
  `password_hash` varchar(100) DEFAULT NULL,
  `salt` varchar(35) DEFAULT NULL,
  `is_superuser` tinyint(1) DEFAULT 0,
  `created` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `mqtt_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `lorawan`.`gateways` (
  `created_at` DATETIME,
  `updated_at` DATETIME,
  `deleted_at` DATETIME,
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(100) DEFAULT NULL,
  `password_hash` varchar(100) DEFAULT NULL,
  `salt` varchar(35) DEFAULT NULL,
  `is_superuser` tinyint(1) DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `mqtt_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `lorawan`.`gateway_acls` (
  `created_at` DATETIME,
  `updated_at` DATETIME,
  `deleted_at` DATETIME,
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(100) NOT NULL,
  `client_id` varchar(100) NOT NULL,
  `action` varchar(9) NOT NULL,
  `permission` varchar(5) NOT NULL,
  `topic` varchar(255) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO `lorawan`.`join_servers`(`username`, `password_hash`, `salt`, `is_superuser`) 
VALUES ('joinserver1', SHA2(concat('123456?aD', 'joinserver1'), 256), 'joinserver1', 1);

INSERT INTO `lorawan`.`gateway_acls`(`username`, `client_id`, `action`, `permission`, `topic`) 
VALUES ('joinserver1', 'joinserver1', 'publish', 'allow', '#');
INSERT INTO `lorawan`.`gateway_acls`(`username`, `client_id`, `action`, `permission`, `topic`) 
VALUES ('joinserver1', 'joinserver1', 'subscribe', 'allow', '#');

-- SELECT password_hash, salt FROM ((SELECT * FROM lorawan.gateways) UNION (SELECT * FROM lorawan.join_servers)) AS user where user.username = 'joinserver1' LIMIT 1
-- DROP TABLE lorawan.join_requests;
-- DROP TABLE lorawan.join_accepts;
-- DROP TABLE lorawan.end_devices;
-- DROP TABLE lorawan.gateways;

-- SELECT * FROM `lorawan`.`join_requests` LIMIT 1000;
-- SELECT * FROM `lorawan`.`join_accepts` LIMIT 1000;
-- SELECT * FROM `lorawan`.`mac_payloads` LIMIT 1000;
-- SELECT * FROM `lorawan`.`end_devices` LIMIT 1000;
-- SELECT * FROM `lorawan`.`gateways` LIMIT 1000;
-- SELECT * FROM `lorawan`.`gateway_acls` LIMIT 1000;

-- UPDATE `lorawan`.`end_devices` SET `dev_nonce` = 0, `join_nonce` = 0 WHERE `id` = 1 OR `id` = 2
-- delete from join_accepts; delete from join_requests; delete from mac_payloads; update end_devices set dev_nonce = 0, join_nonce = 0 where id = 1 or id = 2; 

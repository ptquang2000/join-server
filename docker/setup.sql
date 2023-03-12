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
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(100) DEFAULT NULL,
  `password_hash` varchar(100) DEFAULT NULL,
  `salt` varchar(35) DEFAULT NULL,
  `is_superuser` tinyint(1) DEFAULT 0,
  `created` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `mqtt_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO `lorawan`.`join_servers`(username, password_hash, salt, is_superuser) 
VALUES ('joinserver1', SHA2(concat('123456?aD', 'joinserver1'), 256), 'joinserver1', 1);

-- SELECT password_hash, salt FROM ((SELECT * FROM lorawan.gateways) UNION (SELECT * FROM lorawan.join_servers)) AS user where user.username = 'joinserver1' LIMIT 1

-- DROP TABLE lorawan.join_accepts;
-- DROP TABLE lorawan.join_requests;
-- DROP TABLE lorawan.mac_frames;

-- DELETE FROM lorawan.join_accepts;
-- DELETE FROM lorawan.join_requests;
-- DELETE FROM lorawan.mac_frames;
-- DELETE FROM lorawan.join_requests;DELETE FROM lorawan.join_accepts;DELETE FROM lorawan.mac_frames;

-- SELECT * FROM `lorawan`.`mac_frames` LIMIT 1000;
-- SELECT * FROM `lorawan`.`join_requests` LIMIT 1000;
-- SELECT * FROM `lorawan`.`join_accepts` LIMIT 1000;
-- SELECT * FROM `lorawan`.`end_devices` LIMIT 1000;

-- INSERT INTO `lorawan`.`mac_frames`(frame_id, frame_type) VALUES (1, 1);

GO;
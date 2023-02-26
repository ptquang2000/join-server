-- show query logs
SET GLOBAL log_output = 'TABLE';
SET GLOBAL general_log = 'ON';
SELECT * FROM `mysql`.`general_log` LIMIT 1000;

-- clear query logs
SET GLOBAL general_log= 'OFF';
TRUNCATE table mysql.general_log;

-- EMQX authentication
CREATE TABLE `mqtt_user` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(100) DEFAULT NULL,
  `password_hash` varchar(100) DEFAULT NULL,
  `salt` varchar(35) DEFAULT NULL,
  `is_superuser` tinyint(1) DEFAULT 0,
  `created` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `mqtt_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO mqtt_user(username, password_hash, salt, is_superuser) VALUES ('emqx_u', SHA2(concat('public', 'slat_foo123'), 256), 'slat_foo123', 1);

SELECT password_hash, salt, is_superuser FROM mqtt_user where username = 'emqx_u' LIMIT 1;
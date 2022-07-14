CREATE TABLE IF NOT EXISTS `messages` (
  `id` bigint(20) unsigned NOT NULL,
  `sender` varchar(50) DEFAULT NULL,
  `user_id` bigint(20) DEFAULT NULL,
  `guild_id` bigint(20) DEFAULT NULL,
  `channel_id` bigint(20) DEFAULT NULL,
  `channel` varchar(50) DEFAULT NULL,
  `content` mediumtext NOT NULL DEFAULT '',
  `deleted` bit(1) NOT NULL DEFAULT b'0',
  `reply_id` bigint(20) DEFAULT NULL,
  `aud_create_date` timestamp NOT NULL DEFAULT current_timestamp(),
  `aud_update_date` timestamp NULL DEFAULT NULL ON UPDATE current_timestamp(),
  PRIMARY KEY (`id`),
  KEY `sender` (`sender`),
  KEY `channel` (`channel_id`),
  KEY `guild` (`guild_id`),
  KEY `user_id` (`user_id`),
  FULLTEXT KEY `content` (`content`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `attachments` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `message_id` bigint(20) NOT NULL,
  `url` varchar(4000) DEFAULT NULL,
  `name` varchar(4000) DEFAULT NULL,
  `contents` longblob DEFAULT NULL,
  `is_compressed` bit DEFAULT 0,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=34734 DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `channels` (
  `id` bigint(20) NOT NULL,
  `name` varchar(50) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `edits` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `message_id` bigint(20) unsigned NOT NULL,
  `old_content` blob NOT NULL,
  `create_date` timestamp NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`),
  KEY `FK_edits_messages` (`message_id`),
  CONSTRAINT `FK_edits_messages` FOREIGN KEY (`message_id`) REFERENCES `messages` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=109893 DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `guilds` (
  `id` bigint(20) NOT NULL,
  `name` varchar(50) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `sites` (
  `name` varchar(50) NOT NULL,
  `url` varchar(100) NOT NULL,
  `rss` varchar(100) NOT NULL,
  `servers` varchar(100) NOT NULL,
  `elmahUrl` varchar(100) NOT NULL,
  `channels` varchar(100) NOT NULL,
  `cookie_cobaltsession` varchar(1000) NOT NULL,
  `domain` varchar(100) NOT NULL,
  `max_errors` tinyint(4) NOT NULL DEFAULT 8,
  `period` tinyint(4) NOT NULL DEFAULT 2
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `sites_ignored_errors` (
  `site` varchar(50) NOT NULL,
  `title` varchar(4000) DEFAULT NULL,
  `description` varchar(4000) DEFAULT NULL,
  KEY `site` (`site`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `sites_important_errors` (
  `site` varchar(50) NOT NULL,
  `title` varchar(4000) DEFAULT NULL,
  `description` varchar(4000) DEFAULT NULL,
  KEY `site` (`site`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `sites_timed_out` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `site` varchar(50) NOT NULL,
  `identifier` char(36) DEFAULT NULL,
  `log` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `users` (
  `id` bigint(20) unsigned NOT NULL,
  `name` varchar(100) CHARACTER SET utf8mb4 NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `factoids` (
  `name` varchar(128) NOT NULL,
  `content` varchar(1024) NOT NULL,
  PRIMARY KEY (`name`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `hjts` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) CHARACTER SET utf8mb4 NOT NULL,
  `match_criteria` varchar(128) CHARACTER SET utf8mb4 DEFAULT NULL,
  `description` varchar(1024) DEFAULT NULL,
  `category` varchar(50) DEFAULT NULL,
  `severity` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`),
  KEY `category` (`category`)
) ENGINE=InnoDB AUTO_INCREMENT=189 DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC;

CREATE TABLE IF NOT EXISTS `permissions` (
  `discord_id` varchar(50) NOT NULL,
  `permission` varchar(64) NOT NULL,
  PRIMARY KEY (`discord_id`,`permission`) USING BTREE,
  KEY `discord_id` (`discord_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC;

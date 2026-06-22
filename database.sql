
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

-- 比赛系统数据库迁移
-- 执行: mysql -u root -p onlinepractice < migrations/001_contest_tables.sql

CREATE TABLE IF NOT EXISTS `contest_basic` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` DATETIME(3) DEFAULT NULL,
  `updated_at` DATETIME(3) DEFAULT NULL,
  `deleted_at` DATETIME(3) DEFAULT NULL,
  `identity` VARCHAR(36) NOT NULL,
  `title` VARCHAR(255) NOT NULL,
  `description` TEXT,
  `start_at` DATETIME NOT NULL,
  `end_at` DATETIME NOT NULL,
  `contest_type` TINYINT(1) NOT NULL DEFAULT 1,
  `max_participants` INT(11) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_contest_basic_identity` (`identity`),
  KEY `idx_contest_basic_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `contest_problem` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` DATETIME(3) DEFAULT NULL,
  `updated_at` DATETIME(3) DEFAULT NULL,
  `deleted_at` DATETIME(3) DEFAULT NULL,
  `contest_identity` VARCHAR(36) NOT NULL,
  `problem_identity` VARCHAR(36) NOT NULL,
  `score` INT(11) NOT NULL DEFAULT 100,
  `sort` INT(11) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_contest_problem_contest` (`contest_identity`),
  KEY `idx_contest_problem_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `contest_user` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` DATETIME(3) DEFAULT NULL,
  `updated_at` DATETIME(3) DEFAULT NULL,
  `deleted_at` DATETIME(3) DEFAULT NULL,
  `contest_identity` VARCHAR(36) NOT NULL,
  `user_identity` VARCHAR(36) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_contest_user_contest` (`contest_identity`),
  KEY `idx_contest_user_user` (`user_identity`),
  KEY `idx_contest_user_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `contest_submit` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` DATETIME(3) DEFAULT NULL,
  `updated_at` DATETIME(3) DEFAULT NULL,
  `deleted_at` DATETIME(3) DEFAULT NULL,
  `contest_identity` VARCHAR(36) NOT NULL,
  `problem_identity` VARCHAR(36) NOT NULL,
  `user_identity` VARCHAR(36) NOT NULL,
  `path` VARCHAR(255) NOT NULL,
  `score` INT(11) NOT NULL DEFAULT 0,
  `status` TINYINT(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_contest_submit_contest` (`contest_identity`),
  KEY `idx_contest_submit_user` (`user_identity`),
  KEY `idx_contest_submit_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

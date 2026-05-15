-- ============================================================
-- OnlineJudge 数据库初始化脚本
-- MySQL 8 首次启动时自动执行
-- ============================================================

CREATE DATABASE IF NOT EXISTS onlinepractice
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE onlinepractice;

-- ==================== 用户表 ====================
CREATE TABLE IF NOT EXISTS user_basic (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at  DATETIME(3)       DEFAULT NULL,
    updated_at  DATETIME(3)       DEFAULT NULL,
    deleted_at  DATETIME(3)       DEFAULT NULL,
    identity    VARCHAR(36)       NOT NULL DEFAULT '',
    name        VARCHAR(100)      NOT NULL DEFAULT '',
    password    VARCHAR(60)       NOT NULL DEFAULT '',
    phone       CHAR(11)          NOT NULL DEFAULT '',
    mail        VARCHAR(100)      NOT NULL DEFAULT '',
    pass_num    INT(11)           NOT NULL DEFAULT 0,
    submit_num  INT(11)           NOT NULL DEFAULT 0,
    is_admin    TINYINT(1)        NOT NULL DEFAULT 0,
    UNIQUE INDEX idx_identity (identity),
    INDEX        idx_deleted_at (deleted_at),
    INDEX        idx_name (name),
    INDEX        idx_mail (mail)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==================== 题目表 ====================
CREATE TABLE IF NOT EXISTS problem_basic (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at  DATETIME(3)       DEFAULT NULL,
    updated_at  DATETIME(3)       DEFAULT NULL,
    deleted_at  DATETIME(3)       DEFAULT NULL,
    identity    VARCHAR(36)       NOT NULL DEFAULT '',
    title       VARCHAR(255)      NOT NULL DEFAULT '',
    content     TEXT              NOT NULL,
    max_mem     INT(11)           NOT NULL DEFAULT 0,
    max_runtime INT(11)           NOT NULL DEFAULT 0,
    submit_num  INT(11)           NOT NULL DEFAULT 0,
    pass_num    INT(11)           NOT NULL DEFAULT 0,
    UNIQUE INDEX idx_identity (identity),
    INDEX        idx_deleted_at (deleted_at),
    INDEX        idx_title (title)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==================== 题目分类表 ====================
CREATE TABLE IF NOT EXISTS category_basic (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at  DATETIME(3)       DEFAULT NULL,
    updated_at  DATETIME(3)       DEFAULT NULL,
    deleted_at  DATETIME(3)       DEFAULT NULL,
    identity    VARCHAR(36)       NOT NULL DEFAULT '',
    name        VARCHAR(100)      NOT NULL DEFAULT '',
    parent_id   INT(11)           NOT NULL DEFAULT 0,
    UNIQUE INDEX idx_identity (identity),
    INDEX        idx_deleted_at (deleted_at),
    INDEX        idx_parent_id (parent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==================== 题目-分类关联表 ====================
CREATE TABLE IF NOT EXISTS problem_category (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at  DATETIME(3)       DEFAULT NULL,
    updated_at  DATETIME(3)       DEFAULT NULL,
    deleted_at  DATETIME(3)       DEFAULT NULL,
    problem_id  BIGINT UNSIGNED   NOT NULL DEFAULT 0,
    category_id BIGINT UNSIGNED   NOT NULL DEFAULT 0,
    INDEX        idx_deleted_at (deleted_at),
    INDEX        idx_problem_id (problem_id),
    INDEX        idx_category_id (category_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==================== 测试用例表 ====================
CREATE TABLE IF NOT EXISTS test_case (
    id                BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at        DATETIME(3)       DEFAULT NULL,
    updated_at        DATETIME(3)       DEFAULT NULL,
    deleted_at        DATETIME(3)       DEFAULT NULL,
    identity          VARCHAR(255)      NOT NULL DEFAULT '',
    problem_identity  VARCHAR(255)      NOT NULL DEFAULT '',
    input             TEXT              NOT NULL,
    output            TEXT              NOT NULL,
    INDEX             idx_identity (identity),
    INDEX             idx_deleted_at (deleted_at),
    INDEX             idx_problem_identity (problem_identity)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==================== 提交记录表 ====================
CREATE TABLE IF NOT EXISTS submit_basic (
    id                BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at        DATETIME(3)       DEFAULT NULL,
    updated_at        DATETIME(3)       DEFAULT NULL,
    deleted_at        DATETIME(3)       DEFAULT NULL,
    identity          VARCHAR(36)       NOT NULL DEFAULT '',
    problem_identity  VARCHAR(36)       NOT NULL DEFAULT '',
    user_identity     VARCHAR(36)       NOT NULL DEFAULT '',
    path              VARCHAR(255)      NOT NULL DEFAULT '',
    status            TINYINT(1)        NOT NULL DEFAULT 0,
    INDEX             idx_identity (identity),
    INDEX             idx_deleted_at (deleted_at),
    INDEX             idx_problem_identity (problem_identity),
    INDEX             idx_user_identity (user_identity),
    INDEX             idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

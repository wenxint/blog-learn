-- ============================================================
-- 博客系统初始化脚本
-- 执行：mysql -ublog -p blog < migrations/001_init.sql
-- 顺序：先建被引用的表（users、categories），再建 articles
-- 幂等：全部 IF NOT EXISTS，可重复执行
-- ============================================================

-- 用户表
CREATE TABLE IF NOT EXISTS users (
  id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  username      VARCHAR(50)  NOT NULL COMMENT '登录名',
  password_hash VARCHAR(100) NOT NULL COMMENT 'bcrypt 哈希（60字符）',
  nickname      VARCHAR(50)  NOT NULL DEFAULT '' COMMENT '昵称',
  email         VARCHAR(100) NOT NULL DEFAULT '' COMMENT '邮箱',
  avatar        VARCHAR(255) NOT NULL DEFAULT '' COMMENT '头像URL',
  role          VARCHAR(20)  NOT NULL DEFAULT 'author' COMMENT 'admin/author',
  status        TINYINT      NOT NULL DEFAULT 1 COMMENT '1正常 0禁用',
  created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户表';

-- 分类表
CREATE TABLE IF NOT EXISTS categories (
  id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  name        VARCHAR(50)  NOT NULL COMMENT '分类名',
  slug        VARCHAR(50)  NOT NULL COMMENT 'URL 友好标识',
  description VARCHAR(200) NOT NULL DEFAULT '' COMMENT '描述',
  sort        INT          NOT NULL DEFAULT 0 COMMENT '排序权重，小的在前',
  created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_name (name),
  UNIQUE KEY uk_slug (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='分类表';

-- 标签表
CREATE TABLE IF NOT EXISTS tags (
  id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  name       VARCHAR(50)  NOT NULL COMMENT '标签名',
  slug       VARCHAR(50)  NOT NULL COMMENT 'URL 友好标识',
  created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_name (name),
  UNIQUE KEY uk_slug (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='标签表';

-- 文章表
CREATE TABLE IF NOT EXISTS articles (
  id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  title        VARCHAR(200) NOT NULL COMMENT '标题',
  summary      VARCHAR(500) NOT NULL DEFAULT '' COMMENT '摘要',
  content      LONGTEXT     NOT NULL COMMENT '正文 Markdown',
  cover        VARCHAR(255) NOT NULL DEFAULT '' COMMENT '封面图URL',
  status       TINYINT      NOT NULL DEFAULT 0 COMMENT '1已发布 0草稿',
  is_top       TINYINT      NOT NULL DEFAULT 0 COMMENT '1置顶 0普通',
  view_count   INT          NOT NULL DEFAULT 0 COMMENT '阅读数',
  like_count   INT          NOT NULL DEFAULT 0 COMMENT '点赞数',
  author_id    BIGINT UNSIGNED NOT NULL COMMENT '作者ID → users.id',
  category_id  BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '分类ID → categories.id，0=未分类',
  published_at DATETIME     NULL DEFAULT NULL COMMENT '发布时间，草稿为 NULL',
  created_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  deleted_at   DATETIME     NULL DEFAULT NULL COMMENT '软删除时间，NULL=未删除',
  PRIMARY KEY (id),
  INDEX idx_status_published (status, published_at),
  INDEX idx_category (category_id),
  INDEX idx_author (author_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='文章表';

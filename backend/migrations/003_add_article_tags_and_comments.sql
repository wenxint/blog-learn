-- ============================================================
-- 003：补充文章-标签关联表与评论表
-- 执行：mysql -ublog -p blog < migrations/003_add_article_tags_and_comments.sql
-- 说明：001 已经执行过，不允许修改，新增表只能放在新文件里
-- ============================================================

-- 文章-标签多对多关联表
-- 不设自增主键，用 (article_id, tag_id) 联合主键防止重复关联
CREATE TABLE IF NOT EXISTS article_tags (
  article_id BIGINT UNSIGNED NOT NULL COMMENT '文章ID',
  tag_id     BIGINT UNSIGNED NOT NULL COMMENT '标签ID',
  PRIMARY KEY (article_id, tag_id),
  KEY idx_tag (tag_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='文章-标签关联表';

-- 评论表（支持楼中楼：parent_id=0 是一级评论）
CREATE TABLE IF NOT EXISTS comments (
  id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  article_id BIGINT UNSIGNED NOT NULL COMMENT '文章ID',
  parent_id  BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父评论ID，0=一级评论',
  nickname   VARCHAR(50)  NOT NULL COMMENT '昵称',
  email      VARCHAR(100) NOT NULL DEFAULT '' COMMENT '邮箱（不对外展示）',
  content    VARCHAR(1000) NOT NULL COMMENT '内容',
  status     TINYINT      NOT NULL DEFAULT 0 COMMENT '1已通过 0待审核',
  ip         VARCHAR(45)  NOT NULL DEFAULT '' COMMENT '来源IP',
  created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_article (article_id),
  KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='评论表';

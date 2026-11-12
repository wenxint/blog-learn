-- ============================================================
-- 博客系统种子数据（仅供本地开发/学习）
-- 执行：mysql -ublog -p blog < migrations/002_seed.sql
-- 幂等：用 INSERT IGNORE + 显式 id，重复执行不会产生重复数据
-- 注意：admin 的密码哈希对应明文 "123456"（生产环境绝对不能用）
-- ============================================================

-- 用户：1 个管理员
INSERT IGNORE INTO users
  (id, username, password_hash, nickname, email, role, status)
VALUES
  (1, 'admin', '$2y$10$.iwDzZLeACUS.lWuZS0M4uX10Rt/2hilwwXEaIG/nv8AREabrEYUq', '陶文鑫', 'admin@example.com', 'admin', 1);

-- 分类
INSERT IGNORE INTO categories (id, name, slug, description, sort) VALUES
  (1, '技术笔记', 'tech', '后端、前端、数据库相关的学习记录', 1),
  (2, '生活随笔', 'life', '写点日常和想法', 2);

-- 标签
INSERT IGNORE INTO tags (id, name, slug) VALUES
  (1, 'Go',      'go'),
  (2, 'MySQL',   'mysql'),
  (3, 'Vue',     'vue'),
  (4, '踩坑记录', 'pitfall');

-- 文章：覆盖 已发布/草稿、置顶/不置顶、不同分类、不同阅读数
INSERT IGNORE INTO articles
  (id, title, summary, content, status, is_top, view_count, like_count, author_id, category_id, published_at)
VALUES
  (1, 'Go 从零开始：第一个 HTTP 服务',
      '用标准库 net/http 起一个服务，理解 Handler 到底是什么。',
      '# 第一个 HTTP 服务\n\n这一篇记录用 `net/http` 起服务的过程。\n\n```go\npackage main\n```',
      1, 1, 320, 12, 1, 1, '2026-11-01 10:00:00'),

  (2, 'MySQL 索引到底怎么建',
      '联合索引为什么要讲究顺序？最左前缀原则讲清楚。',
      '# MySQL 索引\n\n联合索引 `(status, published_at)` 能让排序走索引。',
      1, 0, 210, 8, 1, 1, '2026-11-05 14:30:00'),

  (3, 'Vue3 组合式 API 上手笔记',
      'script setup 和以前 Options API 的差别。',
      '# Vue3\n\n`<script setup>` 更简洁。',
      1, 0, 156, 5, 1, 2, '2026-11-08 09:15:00'),

  (4, 'GORM 会用但不懂的那些事',
      'AutoMigrate 到底做了什么，为什么生产不建议直接用。',
      '# GORM\n\nAutoMigrate 会对比结构体自动改表。',
      1, 0, 89, 3, 1, 1, '2026-11-10 20:00:00'),

  (5, '一篇文章还没写完的草稿',
      '这篇还在写，属于草稿状态。',
      '# 草稿\n\n写了一半。',
      0, 0, 0, 0, 1, 2, NULL);

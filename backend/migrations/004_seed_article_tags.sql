-- ============================================================
-- 004：文章-标签关联的种子数据
-- 执行：mysql -ublog -p blog < migrations/004_seed_article_tags.sql
-- 依赖 002 的种子数据（articles 1-5，tags 1-4）
-- 幂等：联合主键 + INSERT IGNORE
-- ============================================================

INSERT IGNORE INTO article_tags (article_id, tag_id) VALUES
  (1, 1),              -- Go 从零开始      -> Go
  (2, 2),              -- MySQL 索引       -> MySQL
  (3, 3),              -- Vue3            -> Vue
  (4, 1), (4, 2), (4, 4),  -- GORM 那篇    -> Go、MySQL、踩坑记录
  (5, 4);              -- 草稿            -> 踩坑记录

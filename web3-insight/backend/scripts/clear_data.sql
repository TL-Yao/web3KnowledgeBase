-- clear_data.sql
-- 清理所有测试数据（保留表结构）
-- WARNING: This will delete ALL data from the database!

-- 禁用外键检查（PostgreSQL）
SET session_replication_role = replica;

-- 清空文章相关表
TRUNCATE TABLE article_versions CASCADE;
TRUNCATE TABLE articles CASCADE;

-- 清空分类表
TRUNCATE TABLE categories CASCADE;

-- 清空聊天消息
TRUNCATE TABLE chat_messages CASCADE;

-- 清空任务队列
TRUNCATE TABLE tasks CASCADE;

-- 清空新闻项
TRUNCATE TABLE news_items CASCADE;

-- 清空数据源（保留结构）
TRUNCATE TABLE data_sources CASCADE;

-- 清空配置（可选，如果想保留配置则注释此行）
-- TRUNCATE TABLE configs CASCADE;

-- 重新启用外键检查
SET session_replication_role = DEFAULT;

-- 验证清理结果
SELECT 'articles' as table_name, COUNT(*) as count FROM articles
UNION ALL
SELECT 'categories', COUNT(*) FROM categories
UNION ALL
SELECT 'chat_messages', COUNT(*) FROM chat_messages
UNION ALL
SELECT 'tasks', COUNT(*) FROM tasks;

-- db/init.sql
CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    duration_hours INT DEFAULT 1,
    priority_score FLOAT DEFAULT 0.0
);

CREATE TABLE dependencies (
    task_id INT REFERENCES tasks(id),
    depends_on_id INT REFERENCES tasks(id),
    PRIMARY KEY (task_id, depends_on_id)
);
-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Обновляем таблицу задач: добавляем владельца
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS user_id INTEGER REFERENCES users(id);
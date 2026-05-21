CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    owner_id INTEGER REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS project_members (
    project_id INTEGER REFERENCES projects(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member',
    PRIMARY KEY (project_id, user_id)
);

CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    opt INT DEFAULT 0,
    real INT DEFAULT 0,
    pess INT DEFAULT 0,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    assignee_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    project_id INTEGER REFERENCES projects(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'todo',
    duration_hours DOUBLE PRECISION DEFAULT 0,
    priority_score DOUBLE PRECISION DEFAULT 0
);

CREATE TABLE IF NOT EXISTS dependencies (
    task_id INTEGER REFERENCES tasks(id) ON DELETE CASCADE,
    depends_on_id INTEGER REFERENCES tasks(id) ON DELETE CASCADE,
    UNIQUE(task_id, depends_on_id)
);

-- Таблица комментариев для задач (Связь с tasks и users)
CREATE TABLE IF NOT EXISTS comments (
    id SERIAL PRIMARY KEY,
    task_id INTEGER REFERENCES tasks(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создаем системного пользователя и дефолтную доску
INSERT INTO users (id, username, password_hash) VALUES (1000, 'system_admin', 'hash') ON CONFLICT (id) DO NOTHING;
INSERT INTO projects (id, name, owner_id) VALUES (1, 'Main Kanban Board', 1000) ON CONFLICT (id) DO NOTHING;
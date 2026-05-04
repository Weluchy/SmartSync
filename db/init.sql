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
    opt FLOAT,
    real FLOAT,
    pess FLOAT,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    project_id INTEGER REFERENCES projects(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'todo',
    duration_hours FLOAT,
    priority_score FLOAT
);

CREATE TABLE IF NOT EXISTS dependencies (
    task_id INTEGER REFERENCES tasks(id) ON DELETE CASCADE,
    depends_on_id INTEGER REFERENCES tasks(id) ON DELETE CASCADE,
    UNIQUE(task_id, depends_on_id)
);

ALTER TABLE tasks 
    ADD COLUMN opt INT DEFAULT 0,
    ADD COLUMN real INT DEFAULT 0,
    ADD COLUMN pess INT DEFAULT 0,
    ADD COLUMN duration_hours DOUBLE PRECISION DEFAULT 0,
    ADD COLUMN priority_score DOUBLE PRECISION DEFAULT 0,
    ADD COLUMN user_id INT REFERENCES users(id);


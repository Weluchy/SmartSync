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
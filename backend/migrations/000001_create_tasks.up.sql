CREATE TABLE tasks (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    title       VARCHAR(255) NOT NULL,
    description TEXT NULL,
    status      ENUM('todo','in_progress','done') NOT NULL DEFAULT 'todo',
    assignee    VARCHAR(255) NULL,
    due_date    DATE NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMP NULL DEFAULT NULL,
    UNIQUE KEY uq_tasks_title (title),
    KEY idx_tasks_status_deleted (status, deleted_at),
    KEY idx_tasks_assignee_deleted (assignee, deleted_at),
    KEY idx_tasks_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

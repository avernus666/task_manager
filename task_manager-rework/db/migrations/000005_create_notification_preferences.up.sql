CREATE TABLE IF NOT EXISTS notification_preferences (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    notification_email VARCHAR(254) NOT NULL,
    email_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    browser_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    task_updates_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    comment_updates_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
INSERT INTO notification_preferences (user_id, notification_email)
SELECT id, email FROM users
ON CONFLICT (user_id) DO NOTHING;

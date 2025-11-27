-- +goose Up
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    data JSONB,
    read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX notifications_user_id_idx ON notifications(user_id);
CREATE INDEX notifications_user_read_idx ON notifications(user_id, read);
CREATE INDEX notifications_created_at_idx ON notifications(created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS notifications_created_at_idx;
DROP INDEX IF EXISTS notifications_user_read_idx;
DROP INDEX IF EXISTS notifications_user_id_idx;
DROP TABLE IF EXISTS notifications;

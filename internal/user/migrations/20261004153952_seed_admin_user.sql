-- +goose Up
INSERT INTO users (email, password_hash, role, status)
VALUES (
        'admin@admin.com',
        '$2a$10$z.cdEJjZzlH9X4vQokP2b.ps5SDBZGzOLef1DF4Xbk.jQRk5IcX12',
        'admin'::user_role,
        'active'::user_status
);

-- +goose Down
DELETE FROM users WHERE email='admin@admin.com';

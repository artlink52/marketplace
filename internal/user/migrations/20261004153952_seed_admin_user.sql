-- +goose Up
INSERT INTO users (email, password_hash, role, status)
VALUES (
        'admin@admin.com',
        '$2a$12$dEqrJnsU7aMpDhQD1Q4iH.HsCIVTLRQflpgFcmkxD2inWFD.x4UTK',
        'admin'::user_role,
        'active'::user_status
);

-- +goose Down
DELETE FROM users WHERE email='admin@admin.com';

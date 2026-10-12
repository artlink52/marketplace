-- +goose Up
INSERT INTO users (email, password_hash, role)
VALUES (
        'seller1@seller.com',
    '$2a$10$9911eGslJMt6k.EDL6OoO.rmYOxQvQ5tajKvdzaLPRSRaahPQz1j2',
        'seller'::user_role
);
INSERT INTO users (email, password_hash, role)
VALUES (
           'seller2@seller.com',
           '$2a$10$9dJxgx3/tdodBIKLdD1uVeit95la/Aw3/f/e636p.nE9Sq64VH25q',
           'seller'::user_role
);
-- +goose Down
DELETE FROM users WHERE email='seller1@seller.com';
DELETE FROM users WHERE email='seller2@seller.com';




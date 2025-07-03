BEGIN;

INSERT INTO authz.credentials (username, password)
VALUES ('admin', '$2a$10$FIzqUNVNN4GfXm3yvJZTCOt2FeAwr7nP8wSM99ZhD83oNUArEXcWi'),
       ('testuser', '$2a$10$FIzqUNVNN4GfXm3yvJZTCOt2FeAwr7nP8wSM99ZhD83oNUArEXcWi'); --hashed with bcrypt

COMMIT;
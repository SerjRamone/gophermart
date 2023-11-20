-- +goose Up
BEGIN;
    
CREATE TABLE IF NOT EXISTS "user" (
    id uuid PRIMARY KEY DEFAULT GEN_RANDOM_UUID(),
    login varchar(155) UNIQUE NOT NULL,
    password varchar(64) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
);

COMMENT ON TABLE "user" IS 'Users';
  
COMMENT ON COLUMN "user".id IS 'Unique user ID';
COMMENT ON COLUMN "user".login IS 'User login';
COMMENT ON COLUMN "user".password IS 'User password';
COMMENT ON COLUMN "user".created_at IS 'Row created date';

COMMIT;

-- +goose Down

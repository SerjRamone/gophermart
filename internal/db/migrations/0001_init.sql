-- +goose Up
BEGIN;
    
-- users ----------------------
CREATE TABLE IF NOT EXISTS "user" (
    id UUID PRIMARY KEY DEFAULT GEN_RANDOM_UUID(),
    login VARCHAR(155) UNIQUE NOT NULL,
    password VARCHAR(64) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE "user" IS 'Users';
  
COMMENT ON COLUMN "user".id IS 'Unique user ID';
COMMENT ON COLUMN "user".login IS 'User login';
COMMENT ON COLUMN "user".password IS 'User password';
COMMENT ON COLUMN "user".created_at IS 'Row created date';

-- order ----------------------
DROP TYPE IF EXISTS order_status;
CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE IF NOT EXISTS "order" (
    id UUID PRIMARY KEY DEFAULT GEN_RANDOM_UUID(),
    user_id UUID NOT NULL REFERENCES "user" (id),
    number BIGINT UNIQUE NOT NULL, 
    accrual DOUBLE PRECISION NOT NULL,
    status order_status NOT NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS user_idx ON "order" (user_id);
CREATE INDEX IF NOT EXISTS uploaded_at_asc_idx ON "order" (uploaded_at ASC);

COMMENT ON TABLE "order" IS 'Orders';

COMMENT ON COLUMN "order".id IS 'Unique order ID';
COMMENT ON COLUMN "order".user_id IS 'User ID';
COMMENT ON COLUMN "order".number IS 'Order number';
COMMENT ON COLUMN "order".accrual IS 'Order accrual sum';
COMMENT ON COLUMN "order".status IS 'Order status';
COMMENT ON COLUMN "order".uploaded_at IS 'Order oploaded date';

-- withdrawal ----------------------
CREATE TABLE IF NOT EXISTS withdrawal (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id UUID NOT NULL REFERENCES "user" (id),
    total DOUBLE PRECISION NOT NULL,
    number BIGINT UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS user_idx ON withdrawal (user_id);

COMMENT ON TABLE withdrawal IS 'Withdrawals';

COMMENT ON COLUMN withdrawal.id IS 'Unique withdrawal ID';
COMMENT ON COLUMN withdrawal.user_id IS 'User ID';
COMMENT ON COLUMN withdrawal.total IS 'Withdrawal total sum';
COMMENT ON COLUMN withdrawal.number IS 'Order number';
COMMENT ON COLUMN withdrawal.created_at IS 'Withdrawal request creation date';

COMMIT;

-- +goose Down

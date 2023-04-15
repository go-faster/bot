-- Modify "telegram_channel_states" table
ALTER TABLE "telegram_channel_states" ALTER COLUMN "pts" SET DEFAULT 0;
-- Modify "telegram_user_states" table
ALTER TABLE "telegram_user_states" ALTER COLUMN "qts" SET DEFAULT 0, ALTER COLUMN "pts" SET DEFAULT 0, ALTER COLUMN "date" SET DEFAULT 0, ALTER COLUMN "seq" SET DEFAULT 0;

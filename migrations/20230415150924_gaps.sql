-- Create "telegram_user_states" table
CREATE TABLE "telegram_user_states" ("id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY, "qts" bigint NOT NULL, "pts" bigint NOT NULL, "date" bigint NOT NULL, "seq" bigint NOT NULL, PRIMARY KEY ("id"));
-- Create "telegram_channel_states" table
CREATE TABLE "telegram_channel_states" ("id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY, "channel_id" bigint NOT NULL, "pts" bigint NOT NULL, "user_id" bigint NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "telegram_channel_states_telegram_user_states_channels" FOREIGN KEY ("user_id") REFERENCES "telegram_user_states" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- Create index "telegramchannelstate_user_id_channel_id" to table: "telegram_channel_states"
CREATE UNIQUE INDEX "telegramchannelstate_user_id_channel_id" ON "telegram_channel_states" ("user_id", "channel_id");
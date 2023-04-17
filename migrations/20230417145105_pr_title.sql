-- Modify "pr_notifications" table
ALTER TABLE "pr_notifications" ADD COLUMN "pull_request_title" character varying NOT NULL DEFAULT '', ADD COLUMN "pull_request_body" character varying NOT NULL DEFAULT '', ADD COLUMN "pull_request_author_login" character varying NOT NULL DEFAULT '';

-- Drop index "check_repo_id_id" from table: "checks"
DROP INDEX "check_repo_id_id";
-- Modify "checks" table
ALTER TABLE "checks" ADD COLUMN "pull_request_id" bigint NOT NULL;
-- Create index "check_repo_id_pull_request_id_id" to table: "checks"
CREATE UNIQUE INDEX "check_repo_id_pull_request_id_id" ON "checks" ("repo_id", "pull_request_id", "id");

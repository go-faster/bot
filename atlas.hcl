// Define an environment named "local"
env "dev" {
  // Define the URL of the Dev Database for this environment
  // See: https://atlasgo.io/concepts/dev-database
  dev = "docker://postgres/15/test?search_path=public"

  # atlas migrate --env dev diff --to ent://internal/ent/schema name
  # https://github.com/ariga/atlas/pull/1582
  # src = "ent://internal/ent/schema"
}

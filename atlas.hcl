// Define an environment named "local"
env "dev" {
  // Define the URL of the Dev Database for this environment
  // See: https://atlasgo.io/concepts/dev-database
  dev = "docker://postgres/15/test?search_path=public"

  # use at least atlas version v0.10.2-7425aae-canary
  src = "ent://internal/ent/schema"
}

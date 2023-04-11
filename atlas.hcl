// Define an environment named "local"
env "dev" {
  // Define the URL of the Dev Database for this environment
  // See: https://atlasgo.io/concepts/dev-database
  dev = "docker://postgres/15/test?search_path=public"
}

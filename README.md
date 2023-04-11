# bot

Bot for go-faster chats and channels, based on [gotd/td](https://github.com/gotd/td).

## Migrations

### Add migration

To add migration named `some-migration-name`:

```console
atlas migrate --env dev diff --to ent://internal/ent/schema some-migration-name
```
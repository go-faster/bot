# bot

Bot for go-faster chats and channels, based on [gotd/td](https://github.com/gotd/td).

## Skip deploy

Add `!skip` to commit message.

## Migrations

### Add migration

To add migration named `some-migration-name`:

```console
atlas migrate --env dev diff some-migration-name
```
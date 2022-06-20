# Absol
Absol is a Discord bot created for many purposes, one of them being helping around in [Minecraft Community Support](https://discord.gg/58Sxm23).

### Runtime

Required environment variables:

- **DISCORD_TOKEN**: Bot discord token to log into Discord
- **DATABASE_DIALECT**: Dialect of the database (mysql only, default mysql)
- **DATABASE_USER**: Database username (default discord)
- **DATABASE_PASS**: Database password (default discord)
- **DATABASE_HOST**: Database host (default empty, equals localhost)
- **DATABASE_DB**: Database name to use (default discord)


All environment variables may add _FILE to the end to read from a Docker secret path. The path to the secret should be set as the value.
Example: DISCORD_TOKEN_FILE=/run/secrets/discordtoken will read the value from the discordtoken secret

### Docker Build
Run `docker build -t absol .` and if it dosent find any errors you should see `Successfully built`.
Then you can start the docker container by running `docker run -it absol`

### Go build
run `go build -o absol -tags=modules.all,databases.all -v github.com/LordRalex/absol`, then `./absol`

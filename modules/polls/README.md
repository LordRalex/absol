# Polls

## Required ENV settings:

- APP_ID - application id to create slash commands from
- POLLS_GUILDS - guilds (separated by ;) to enable command in
- DATABASE - Fully defined MySQL connection string (refer to https://github.com/go-sql-driver/mysql#dsn-data-source-name for full information)


## Database

Due to internal limitations, this only runs on MySQL/MariaDB.

Initial startup will create the tables needed. The user defined should have the following rights to the database (it is recommended to create a new database that this user only has access to)

- SELECT
- UPDATE
- CREATE
- ALTER

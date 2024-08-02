## Intro

This repository contains a simple CLI tool to fetch and monitor public github repositories.
It provides simple and interactive cli controls to fetch repository metadata, fetch commits, and see top contributors. Additionally it provides a way to fetch commits since a specific date. 

## Dependancies
The app uses postgres as a persistent so you will need have postgres running and configure 
the environment variables.
- DB_HOST = < pg-address >
- DB_PORT = < pg-port >
- DB_NAME =  < name-of-database >
- DB_USERNAME =  < pg-username >
- DB_PASSWORD =  < pg-password >

If you don't have postgres. You can use the simple docker compose file in this repo to pull and run postgres easily.

### Extra
By default, it refreshes all the repo data every hour. But you can use the INTERVAL env variable to configure the amount of time it waits between refreshes.
- INTERVAL = <#HOURS>

The app creates two log files app.log and error.log.
You'll find errors in error.log and other app logs in the app.log file.

You can either build and run the app and use 
```go run *.go``` to run and test the app
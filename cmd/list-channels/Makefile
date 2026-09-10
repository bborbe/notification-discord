include ../../../../../Makefile.variables
include ../../../../../Makefile.env
include ../../../../../common.env
include ../../../../../local.env

run:
	@go mod tidy && go run -mod=mod main.go \
	-sentry-dsn="$(shell teamvault-cli url --teamvault-key=${SENTRY_DSN_KEY})" \
	-discord-server-id="${CORE_DISCORD_SERVER_ID}" \
	-discord-token="$(shell teamvault-cli password --teamvault-key=${CORE_DISCORD_KEY})" \
	-v=2

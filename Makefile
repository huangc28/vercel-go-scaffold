## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

#==============================================================
# Setup
#==============================================================

## setup: initial setup for new project
.PHONY: setup
setup:
	@./scripts/setup.sh

#==============================================================
# Development
#==============================================================

## vet: run go vet
.PHONY: vet
vet:
	go vet ./...

## test: run all tests
.PHONY: test
test:
	go test ./...

## test/coverage: run tests with coverage
.PHONY: test/coverage
test/coverage:
	go test -cover ./...

## sqlc/generate: generate go code from sql queries
.PHONY: sqlc/generate
sqlc/generate:
	sqlc generate

## supabase/db-dump: dump the local db to a file
.PHONY: supabase/db-dump
supabase/db-dump:
	supabase db dump --file supabase/schemas/schema.sql

#==============================================================
# Vercel
#==============================================================

## start/vercel: start the vercel dev server
.PHONY: start/vercel
start/vercel:
	vercel dev --debug --listen 3008
	@echo "Vercel dev server started on port 3008"

## deploy/vercel/preview: deploy the preview version of the vercel app
.PHONY: deploy/vercel/preview
deploy/vercel/preview:
	vercel deploy

## deploy/vercel/prod: deploy the production version of the vercel app
.PHONY: deploy/vercel/prod
deploy/vercel/prod:
	vercel deploy --prod --debug

#==============================================================
# Telegram bot
#==============================================================
.PHONY: telegram/set-webhook/local
telegram/set-webhook/local:
	curl "https://api.telegram.org/bot7944292479:AAEnzoQ_YGmj5qLWFxfabMweto_NDmg-u0c/setWebhook?url=https://da2c-111-241-103-226.ngrok-free.app/v1/webhooks/telegram" -X POST
##@ Dev stack (docker-compose.yml)

.PHONY: dev-up dev-down dev-restart dev-build dev-build-clean dev-logs dev-ps dev-attach dev-shell

dev-up: ## Start the dev stack (build if needed), detached
	$(DC) up -d --build

dev-attach: ## Start-or-attach to the dev stack in the foreground (Ctrl-C stops it; use dev-logs to just watch)
	$(DC) up

dev-down: ## Stop the dev stack (keeps postgres-data & named volumes)
	$(DC) down

dev-restart: ## Restart the app service (re-reads .env, re-runs air)
	$(DC) restart $(APP)

dev-build: ## Build the dev app image (uses cache)
	$(DC) build $(APP)

dev-build-clean: ## Rebuild the dev app image with no cache
	$(DC) build --no-cache $(APP)

dev-logs: ## Tail dev logs — one service with `make dev-logs S=app`
	$(DC) logs -f $(S)

dev-shell: ## Open a bash shell inside the running dev app container
	$(DC) exec $(APP) bash

dev-ps: ## Show dev container status
	$(DC) ps

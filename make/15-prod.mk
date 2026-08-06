##@ Prod stack (docker-compose.prod.yml)

.PHONY: prod-up prod-down prod-build prod-build-clean prod-logs prod-ps prod-attach prod-shell

prod-up: ## Start the prod stack (single embedded binary), detached
	$(DC_PROD) up -d --build

prod-attach: ## Start-or-attach to the prod stack in the foreground (Ctrl-C stops it; use prod-logs to just watch)
	$(DC_PROD) up

prod-down: ## Stop the prod stack
	$(DC_PROD) down

prod-build: ## Build the prod image (uses cache)
	$(DC_PROD) build

prod-build-clean: ## Rebuild the prod image with no cache
	$(DC_PROD) build --no-cache

prod-logs: ## Tail prod logs — one service with `make prod-logs S=app`
	$(DC_PROD) logs -f $(S)

prod-shell: ## Open a bash shell inside the running prod app container
	$(DC_PROD) exec $(APP) bash

prod-ps: ## Show prod container status
	$(DC_PROD) ps

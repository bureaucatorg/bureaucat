##@ Database

.PHONY: migrate seed db-shell

migrate: ## Run DB migrations (inside the app container)
	$(DC) exec -T $(APP) $(BIN) migrate up

seed: ## Seed rich demo data (random users/workspace/projects/cycles/modules/tasks). RESETS app data.
	$(PYTHON) tools/seed.py

db-shell: ## Open an interactive psql shell in the postgres container
	$(DC) exec $(PG) psql -U bureaucat -d bureaucat

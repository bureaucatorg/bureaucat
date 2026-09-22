##@ Cleanup

.PHONY: clean nuke

clean: ## Stop dev + prod stacks and drop their named volumes (keeps data dirs)
	-$(DC) down -v
	-$(DC_PROD) down -v
	@echo "stopped + named volumes removed. Use 'make nuke' to also remove built images + data."

nuke: ## Remove THIS repo's containers, volumes, networks, built images + local data
	-$(DC) down -v --remove-orphans --rmi local
	-$(DC_PROD) down -v --remove-orphans --rmi local
	-docker run --rm -v "$$PWD":/w -w /w alpine sh -c 'rm -rf garage postgres-data'
	@echo "nuked: bureaucat dev+prod containers/volumes/networks/built-images + local data (shared base images untouched)."

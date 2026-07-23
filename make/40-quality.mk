##@ Quality

.PHONY: fmt tidy

fmt: ## Format Go code (go fmt, inside the app container)
	$(DC) exec -T $(APP) go fmt ./...

tidy: ## Tidy go.mod / go.sum (inside the app container)
	$(DC) exec -T $(APP) go mod tidy

##@ Help

.DEFAULT_GOAL := help

.PHONY: help

help: ## Show this help, grouped by section
	@awk 'BEGIN { FS = ":.*##"; n = 0 } \
		/^##@/ { typ[n] = "H"; txt[n] = substr($$0, 5); n++; next } \
		/^[a-zA-Z0-9_-]+:.*##/ { \
			d = $$2; sub(/^[ \t]+/, "", d); \
			typ[n] = "T"; key[n] = $$1; txt[n] = d; n++; \
			if (length($$1) > w) w = length($$1) } \
		END { for (i = 0; i < n; i++) { \
			if (typ[i] == "H") printf "\n\033[1m%s\033[0m\n", txt[i]; \
			else printf "  \033[36m%-*s\033[0m  %s\n", w, key[i], txt[i] } \
			print "" }' $(MAKEFILE_LIST)

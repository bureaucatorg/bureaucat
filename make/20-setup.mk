##@ Setup

.PHONY: setup bootstrap dev-bootstrap prod-bootstrap env garage-config _stop-all

setup: env garage-config ## One-time local setup: create .env + garage config
	@echo "✓ setup done — next: make dev-bootstrap"

env: $(ENV_FILE) ## Create .env from .env.example (if missing)
$(ENV_FILE):
	@cp $(ENV_EXAMPLE) $(ENV_FILE) && echo "→ created $(ENV_FILE)"

garage-config: $(GARAGE_TOML) ## Generate garage/garage.toml with fresh secrets (if missing)
$(GARAGE_TOML):
	@mkdir -p garage/meta garage/data
	@rpc=$$(openssl rand -hex 32); adm=$$(openssl rand -hex 32); \
	{ \
	  echo 'metadata_dir = "/var/lib/garage/meta"'; \
	  echo 'data_dir = "/var/lib/garage/data"'; \
	  echo 'db_engine = "sqlite"'; \
	  echo ''; \
	  echo 'replication_factor = 1'; \
	  echo ''; \
	  echo 'rpc_bind_addr = "[::]:3901"'; \
	  echo 'rpc_public_addr = "127.0.0.1:3901"'; \
	  echo "rpc_secret = \"$$rpc\""; \
	  echo ''; \
	  echo '[s3_api]'; \
	  echo 's3_region = "garage"'; \
	  echo 'api_bind_addr = "[::]:3900"'; \
	  echo 'root_domain = ".s3.garage.localhost"'; \
	  echo ''; \
	  echo '[s3_web]'; \
	  echo 'bind_addr = "[::]:3902"'; \
	  echo 'root_domain = ".web.garage.localhost"'; \
	  echo 'index = "index.html"'; \
	  echo ''; \
	  echo '[admin]'; \
	  echo 'api_bind_addr = "[::]:3903"'; \
	  echo "admin_token = \"$$adm\""; \
	} > $(GARAGE_TOML); \
	echo "→ generated $(GARAGE_TOML) (fresh secrets)"

# Stop + remove BOTH stacks so dev/prod can't clash on ports or the shared
# bureaucat-garage container name.
_stop-all:
	@echo "→ stopping any existing dev/prod stacks..."
	-$(DC) down --remove-orphans
	-$(DC_PROD) down --remove-orphans

bootstrap: dev-bootstrap ## Alias for dev-bootstrap

dev-bootstrap: setup _stop-all ## Turnkey DEV: fresh stack → S3 bucket+keys → app → migrate → seed
	@echo "→ starting postgres + garage..."
	$(DC) up -d --build $(PG) garage
	@echo "→ waiting for garage to become healthy..."
	@until [ "$$(docker inspect -f '{{.State.Health.Status}}' bureaucat-garage 2>/dev/null)" = healthy ]; do sleep 2; done
	@echo "→ creating S3 bucket + access key, wiring them into .env..."
	@./garage-init.sh
	@echo "→ starting the app..."
	$(DC) up -d --build
	@echo "→ waiting for the API to answer..."
	@until curl -sf $(HEALTH_URL) >/dev/null 2>&1; do sleep 2; done
	@$(MAKE) --no-print-directory migrate
	@$(MAKE) --no-print-directory seed
	@echo ""
	@echo "✅ Dev is up at $(APP_URL)  (demo login: demo@gmail.com / Passw0rd!)"

prod-bootstrap: setup _stop-all ## Turnkey PROD: fresh stack → S3 bucket+keys → app → seed
	@echo "→ starting postgres + garage..."
	$(DC_PROD) up -d --build $(PG) garage
	@echo "→ waiting for garage to become healthy..."
	@until [ "$$(docker inspect -f '{{.State.Health.Status}}' bureaucat-garage 2>/dev/null)" = healthy ]; do sleep 2; done
	@echo "→ creating S3 bucket + access key, wiring them into .env..."
	@./garage-init.sh
	@echo "→ building + starting the app..."
	$(DC_PROD) up -d --build
	@echo "→ waiting for the API to answer..."
	@until curl -sf $(HEALTH_URL) >/dev/null 2>&1; do sleep 2; done
	@PG_CONTAINER=bureaucat-prod-postgres-1 $(PYTHON) tools/seed.py
	@echo ""
	@echo "✅ Prod is up at $(APP_URL)  (demo login: demo@gmail.com / Passw0rd!)"

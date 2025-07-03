.PHONY: fmt up down logs ps clean migrate-create generate build-codegen start

GENERATOR_IMAGE := oapi-codegen-v2
GENERATOR_DOCKERFILE := oapi-codegen.Dockerfile
OPENAPI_FILE := openapi.yaml
OUT_DIR := internal/api
CODEGEN_CONFIG := oapi-codegen.cfg.yaml

COMPOSE_CMD := docker compose

# --- Go Tooling ---
fmt:
	@echo "✨ Formatting Go code..."
	go fmt ./...

# --- Docker Compose Commands ---
up:
	@echo "🚀 Starting services with Docker Compose..."
	$(COMPOSE_CMD) up --build -d

down:
	@echo "🛑 Stopping services..."
	$(COMPOSE_CMD) down

logs:
	@echo "📜 Tailing logs..."
	$(COMPOSE_CMD) logs -f

ps:
	@echo "📋 Listing running services..."
	$(COMPOSE_CMD) ps

clean:
	@echo "🧹 Cleaning up volumes and generated code..."
	$(COMPOSE_CMD) down -v --rmi all --remove-orphans
	rm -rf $(OUT_DIR)

migrate-create:
	@read -p "Enter migration name (e.g., add_users_table): " name; \
	docker run --rm -v "${PWD}/migrations:/migrations" migrate/migrate create -ext sql -dir /migrations $$name
	@echo "✅ Created new migration file in ./migrations/"


build-codegen:
	@echo "🛠️  Building codegen Docker image..."
	docker build \
		-t $(GENERATOR_IMAGE) \
		-f $(GENERATOR_DOCKERFILE) .

generate: build-codegen
	@echo "🧬 Generating API code using 'gin-server' generator..."
	@mkdir -p $(OUT_DIR)
	docker run --rm \
		-v "${PWD}:/work" \
		-w /work \
		--user "$$(id -u):$$(id -g)" \
		$(GENERATOR_IMAGE) \
		--config $(CODEGEN_CONFIG) $(OPENAPI_FILE)
	@echo "✅ API code generated successfully."

unit-test:
	@echo "🏃 Running unit tests…"
	go test ./test/unit... -short

integration-test:
	@echo "🏃 Running integration tests…"
	go test ./test/itest/... -timeout 60s

test: unit-test integration-test
	@echo "✅ All tests passed!"

start: clean generate up
	@echo "🎉 Application started successfully! Tailing logs now."
	@$(MAKE) logs
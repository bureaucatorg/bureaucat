.PHONY: dev build build-frontend clean

# Development mode - starts both Go server and Nuxt dev server
dev:
	go run ./cmd/bureaucat serve --dev

# Build frontend and copy to embed directory
build-frontend:
	cd web && bun run build
	rm -rf cmd/bureaucat/dist/*
	cp -r web/.output/public/* cmd/bureaucat/dist/

# Build production binary with embedded frontend
build: build-frontend
	go build -o bureaucat ./cmd/bureaucat

# Clean build artifacts
clean:
	rm -rf bureaucat
	rm -rf web/.output
	rm -rf cmd/bureaucat/dist/*
	touch cmd/bureaucat/dist/.gitkeep

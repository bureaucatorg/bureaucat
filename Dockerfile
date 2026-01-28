# Build frontend
FROM oven/bun:1.3.7-slim AS frontend
WORKDIR /app/web
COPY web/package.json web/bun.lock ./
RUN bun install --frozen-lockfile
COPY web .
RUN bun run generate

# Build backend
FROM golang:1.25.6-trixie AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Copy Nuxt generated static files to embed location
COPY --from=frontend /app/web/.output/public ./cmd/bureaucat/dist
# Copy migrations to embed location
RUN cp -r migrations/* ./cmd/bureaucat/migrations/
RUN go build -o bureaucat ./cmd/bureaucat

# Runtime
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=backend /app/bureaucat .
EXPOSE 1323
CMD ["./bureaucat", "serve"]

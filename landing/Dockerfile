# Build context is the repo root (shared by docker-compose and goreleaser).
FROM oven/bun:1.3.7-slim AS build
WORKDIR /app
COPY landing/package.json landing/bun.lock ./
RUN bun install --frozen-lockfile
COPY landing/index.html landing/vite.config.ts landing/tsconfig.json ./
COPY landing/scripts ./scripts
COPY landing/src ./src
COPY landing/public ./public
RUN bun run build

FROM nginxinc/nginx-unprivileged:1.29-alpine
ENV APP_URL=""
COPY landing/nginx/default.conf.template /etc/nginx/templates/default.conf.template
COPY landing/nginx/security-headers.conf /etc/nginx/snippets/security-headers.conf
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 8080

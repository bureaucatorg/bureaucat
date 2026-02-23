# Bureaucat - Project Guide for Claude

## Critical Rules

1. **No build/codegen command execution** - Never execute `bun`, `go`, or `sqlc` commands. Always ask the user to run these commands manually.
2. **Static site only** - Nuxt is configured for static site generation (SSR disabled). All data is hydrated via APIs from the Go Echo server. Do not create server-side rendered pages.
3. **API prefix** - All Go API routes must start with `/api/v1/`
4. **Admin routes** - Admin endpoints use `/api/v1/admin/*` in Go and `/admin/*` in Nuxt
5. **UI framework** - Always use shadcn-vue components with Tailwind CSS v4. Write all frontend code in TypeScript.

## Project Overview

Bureaucat is a full-stack approval workflow application with:
- **Backend:** Go 1.25.6 with Echo v5 framework
- **Frontend:** Nuxt 4.3.0 (Vue 3.5) with TypeScript
- **Database:** PostgreSQL 18.1 with sqlc for type-safe queries
- **Styling:** Tailwind CSS v4 + shadcn-vue (new-york style)
- **Package Manager:** Bun (frontend only)
- **Deployment:** Single binary with embedded frontend assets

## Directory Structure

```
bureaucat/
├── cmd/bureaucat/           # Application entry point
│   ├── main.go              # CLI entry, embedded assets
│   ├── embed.go             # Go embed directives
│   ├── dist/                # Production frontend (embedded)
│   └── migrations/          # Production migrations (embedded)
│
├── internal/                # Core application code
│   ├── cli/                 # CLI commands (serve, migrate)
│   ├── server/              # HTTP server, routes, proxy/static
│   ├── handlers/            # HTTP handlers (auth.go, admin.go)
│   ├── auth/                # JWT, password, middleware
│   ├── store/               # sqlc-generated database layer
│   └── database/            # Migration management
│
├── migrations/              # SQL migration files (####_name.up/down.sql)
├── queries/                 # sqlc query definitions (*.sql)
│
├── web/                     # Nuxt frontend
│   ├── app/
│   │   ├── pages/           # File-based routing
│   │   ├── components/      # Vue components
│   │   │   └── ui/          # shadcn-vue components
│   │   ├── composables/     # useAuth.ts, useAdmin.ts
│   │   ├── middleware/      # auth.ts, guest.ts, admin.ts
│   │   ├── plugins/         # auth.client.ts
│   │   ├── utils/           # validators.ts
│   │   └── lib/             # utils.ts (cn helper)
│   ├── nuxt.config.ts
│   └── components.json      # shadcn-nuxt config
│
├── Dockerfile               # Production multi-stage build
├── docker-compose.yml       # Development environment
├── docker-compose.prod.yml  # Production environment
├── sqlc.yaml                # sqlc configuration
└── .air.toml                # Hot reload configuration
```

## API Routes

### Public Routes
```
POST /api/v1/signup           # User registration
POST /api/v1/signin           # User login
POST /api/v1/token_refresh    # Refresh access token
POST /api/v1/logout           # Logout (revoke tokens)
GET  /api/v1/health           # Health check
GET  /api/v1/ht/              # Detailed health check
```

### Protected Routes (requires Bearer token)
```
GET  /api/v1/me               # Get current user
```

### Admin Routes (requires Bearer token + admin role)
```
GET    /api/v1/admin/users           # List users (paginated)
POST   /api/v1/admin/users           # Create user
DELETE /api/v1/admin/users/:id       # Delete user
GET    /api/v1/admin/tokens          # List active refresh tokens
DELETE /api/v1/admin/tokens/:id      # Revoke token
DELETE /api/v1/admin/tokens/expired  # Clean expired tokens
```

## Authentication

- **Access Token:** JWT (HS256), 5 min expiry, sent as `Authorization: Bearer <token>`
- **Refresh Token:** Random 32-byte, SHA-256 hashed in DB, 7-day expiry, httpOnly cookie
- **Token Rotation:** Old refresh token revoked on each refresh
- **Password:** bcrypt with cost 12

## Frontend Patterns

### Adding a New Page
1. Create `web/app/pages/pagename.vue`
2. Use `definePageMeta({ middleware: ['auth'] })` for protected pages
3. Fetch data from API using composables

### Using shadcn Components
```vue
<script setup lang="ts">
import { Button } from '~/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '~/components/ui/card'
</script>
```

### Making API Calls
```typescript
const { getAuthHeader, isAuthenticated } = useAuth()

const response = await fetch('/api/v1/endpoint', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    ...getAuthHeader()
  },
  credentials: 'include',
  body: JSON.stringify(data)
})
```

## Backend Patterns

### Adding a New Handler
1. Create handler function in `internal/handlers/`
2. Register route in `internal/server/routes.go`
3. Use middleware: `authMiddleware` for protected, add `adminMiddleware` for admin

### Handler Structure
```go
func (h *Handler) MyEndpoint(c echo.Context) error {
    userID := c.Request().Header.Get("X-User-ID")
    // ... handle request
    return c.JSON(http.StatusOK, response)
}
```

### Adding Database Queries
1. Write SQL in `queries/*.sql` with sqlc annotations
2. Ask user to run: `sqlc generate`
3. Use generated methods from `internal/store/`

### Adding Migrations
1. Create `migrations/XXXX_name.up.sql` and `migrations/XXXX_name.down.sql`
2. Ask user to run: `./bureaucat migrate up`

## Development

```bash
# Start development environment (user runs this)
make dev
# or
docker-compose up
```

- Go API: http://localhost:1341
- Nuxt dev server: http://localhost:3041
- Hot reload enabled for both

## Production Build

```bash
# Build for production (user runs this)
make build
# or
docker-compose -f docker-compose.prod.yml up --build
```

Produces a single binary with embedded frontend and migrations.

## Environment Variables

Required in `.env`:
```
JWT_SECRET=your-secret-key-minimum-32-chars
DATABASE_URL=postgres://user:pass@host:5432/dbname?sslmode=disable
```

Optional:
```
ACCESS_TOKEN_EXPIRY_MINS=5
REFRESH_TOKEN_EXPIRY_DAYS=7
```

## Key Files Reference

| File | Purpose |
|------|---------|
| `internal/server/routes.go` | All API route registration |
| `internal/handlers/auth.go` | Authentication endpoints |
| `internal/handlers/admin.go` | Admin endpoints |
| `internal/auth/middleware.go` | Auth middleware |
| `web/app/composables/useAuth.ts` | Frontend auth state |
| `web/app/composables/useAdmin.ts` | Admin API calls |
| `queries/auth.sql` | Database query definitions |
| `sqlc.yaml` | sqlc code generation config |

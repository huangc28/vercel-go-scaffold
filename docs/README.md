# 🤖 AI Assistant Documentation Index

Welcome! This project includes comprehensive documentation designed specifically for AI assistants (GitHub Copilot, Cursor, etc.) to understand the codebase and help developers effectively.

## 📋 Documentation Files

### Primary AI Context Files
1. **`.cursorrules`** (Root) - Primary configuration file that Cursor automatically loads
2. **`docs/AI_DEVELOPMENT_GUIDE.md`** - Comprehensive development guide with examples
3. **`docs/PROJECT_CONVENTIONS.md`** - Coding standards and best practices
4. **`docs/AI_QUICK_REFERENCE.md`** - Quick commands and structure overview

### IDE Configuration
- **`.vscode/settings.json`** - VS Code settings optimized for Go development and Copilot

## 🎯 Quick Start for AI Assistants

### Project Type
**Go + Chi + Vercel + Supabase** serverless API starter

### Key Technologies
- **Backend**: Go 1.21+
- **Router**: Chi v5
- **Database**: PostgreSQL (Supabase) with SQLC
- **Deployment**: Vercel serverless functions
- **DI Framework**: Uber FX
- **Logging**: Zap

### Import Path
`github/huangc28/kikichoice-be` (update in `go.mod` for new projects)

## 🏗️ Architecture Overview

```
Entry Points (Vercel) → Handlers (Business Logic) → Database (SQLC)
```

### Core Directories
- `api/go/entries/` - Vercel serverless function endpoints
- `api/go/_internal/handlers/` - Business logic handlers
- `api/go/_internal/db/` - Database layer with SQLC
- `api/go/_internal/pkg/` - Shared utilities

## 🔧 Development Workflow

### Adding New API Endpoint
1. Create handler in `api/go/_internal/handlers/feature/`
2. Create entry point in `api/go/entries/feature.go`
3. Add routing to `vercel.json`
4. Add SQL queries to `queries.sql` (if needed)
5. Run `make sqlc/generate`

### Essential Commands
```bash
make sqlc/generate        # Generate database code
make start/vercel         # Start dev server
make test                 # Run tests
make deploy/vercel/preview # Deploy preview
```

## 📝 Code Patterns

### Handler Template
```go
func NewHandler(p HandlerParams) *Handler
func (h *Handler) RegisterRoutes(r *chi.Mux)
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request)
```

### Response Pattern
```go
// Success
render.ChiJSON(w, r, data)

// Error
render.ChiErr(w, r, err, "ERROR_CODE", 
    render.WithStatusCode(http.StatusBadRequest))
```

### Dependency Injection
```go
type HandlerParams struct {
    fx.In
    Logger *zap.SugaredLogger
    Config *configs.Config
    // other dependencies
}
```

## 🗄️ Database Operations

### Query Definition (`queries.sql`)
```sql
-- name: GetUser :one
-- name: ListUsers :many
-- name: CreateUser :one
-- name: UpdateUser :one
-- name: DeleteUser :exec
```

### Usage in Handlers
```go
user, err := h.queries.GetUser(ctx, userID)
users, err := h.queries.ListUsers(ctx, db.ListUsersParams{
    Limit: 10,
    Offset: 0,
})
```

## 🛠️ Common Tasks

### When Creating New Features
- Follow the handler pattern in `docs/AI_DEVELOPMENT_GUIDE.md`
- Use error codes from `err_codes.go`
- Implement proper logging and error handling
- Add database queries if needed
- Create corresponding entry points

### When Debugging
- Check logs with structured logging
- Use VS Code debugger with Go extension
- Test endpoints locally at `http://localhost:3008`

## 📚 Full Documentation

For complete details, see:
- **`.cursorrules`** - Comprehensive project context
- **`docs/AI_DEVELOPMENT_GUIDE.md`** - Detailed examples and patterns
- **`docs/PROJECT_CONVENTIONS.md`** - Coding standards
- **`README.md`** - Project setup and overview

---

This documentation ensures AI assistants have full context to help developers build consistent, maintainable APIs following established patterns.

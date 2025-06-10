# AI Assistant Development Guide

This comprehensive guide helps AI assistants understand the Go Chi Vercel starter project structure, conventions, and best practices.

## 🏗️ Project Architecture

This starter follows a **layered serverless architecture** with these core principles:

### Architecture Layers
1. **Entry Points** (`api/go/entries/`) - Vercel serverless function handlers
2. **Handlers** (`api/go/_internal/handlers/`) - Business logic controllers
3. **Database** (`api/go/_internal/db/`) - Data access layer with SQLC
4. **Infrastructure** (`api/go/_internal/pkg/`) - Shared utilities and helpers

### Dependency Injection
- Uses **Uber FX** for dependency injection
- All dependencies declared with `fx.In` struct tags
- Handlers receive logger, config, and other services via DI

## 📁 Detailed Directory Structure

```
go-chi-vercel-starter/
├── api/go/
│   ├── entries/                    # 🚀 Vercel Function Endpoints
│   │   ├── hello.go               # Example endpoint handler
│   │   └── your-endpoint.go       # Your new endpoints here
│   └── _internal/                 # 🔒 Internal Application Code
│       ├── configs/               # ⚙️ Configuration Management
│       │   └── core.go           # Config loading and validation
│       ├── db/                    # 🗄️ Database Layer
│       │   ├── core.go           # Database connection setup
│       │   ├── tx.go             # Transaction helpers
│       │   ├── models.go         # 🤖 SQLC Generated Models
│       │   ├── queries.sql.go    # 🤖 SQLC Generated Queries
│       │   ├── db.go             # 🤖 SQLC Generated Interfaces
│       │   └── sqlc_queries/     # 📝 SQL Query Definitions
│       │       └── queries.sql   # Raw SQL queries for SQLC
│       ├── fx/                    # 🔧 Dependency Injection
│       │   └── options.go        # FX module options
│       ├── handlers/              # 🎯 Business Logic Handlers
│       │   └── hello/            # Example handler package
│       │       ├── handler.go    # Handler implementation
│       │       └── err_codes.go  # Error constants
│       ├── middlewares/           # 🛡️ HTTP Middlewares
│       │   ├── response_logger.go # Response logging middleware
│       │   └── err_codes.go      # Middleware error constants
│       ├── pkg/                   # 📦 Shared Packages
│       │   ├── logger/           # Logging utilities
│       │   └── render/           # Response rendering utilities
│       └── router/               # 🛣️ Router Configuration
│           ├── core.go           # Main router setup
│           └── fx/               # Router FX options
├── docs/                          # 📚 Documentation
├── supabase/                      # 🏢 Database Schema & Migrations
│   ├── migrations/               # Database migration files
│   └── schemas/                  # SQL schema definitions
├── scripts/                       # 🔧 Setup and Utility Scripts
└── [config files]               # Various config files
```

## 🎯 Handler Development Pattern

### Complete Handler Example

```go
// api/go/_internal/handlers/users/handler.go
package users

import (
    "encoding/json"
    "net/http"
    "strconv"
    
    "github.com/go-chi/chi/v5"
    "github/huangc28/kikichoice-be/api/go/_internal/configs"
    "github/huangc28/kikichoice-be/api/go/_internal/db"
    "github/huangc28/kikichoice-be/api/go/_internal/pkg/render"
    "go.uber.org/fx"
    "go.uber.org/zap"
)

// UserResponse represents a user API response
type UserResponse struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// CreateUserRequest represents request payload for creating users
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

// UsersHandler handles user-related requests
type UsersHandler struct {
    logger  *zap.SugaredLogger
    config  *configs.Config
    queries *db.Queries
}

// UsersHandlerParams defines dependencies for users handler
type UsersHandlerParams struct {
    fx.In
    
    Logger  *zap.SugaredLogger
    Config  *configs.Config
    Queries *db.Queries
}

// NewUsersHandler creates a new users handler instance
func NewUsersHandler(p UsersHandlerParams) *UsersHandler {
    return &UsersHandler{
        logger:  p.Logger,
        config:  p.Config,
        queries: p.Queries,
    }
}

// RegisterRoutes registers user routes with the chi router
func (h *UsersHandler) RegisterRoutes(r *chi.Mux) {
    r.Route("/users", func(r chi.Router) {
        r.Get("/", h.ListUsers)
        r.Post("/", h.CreateUser)
        r.Get("/{id}", h.GetUser)
        r.Put("/{id}", h.UpdateUser)
        r.Delete("/{id}", h.DeleteUser)
    })
}

// GetUser retrieves a single user by ID
func (h *UsersHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    h.logger.Info("Getting user")
    
    // Extract ID from URL parameter
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        h.logger.Errorw("Invalid user ID", "id", idStr, "error", err)
        render.ChiErr(w, r, err, InvalidUserID, 
            render.WithStatusCode(http.StatusBadRequest))
        return
    }
    
    // Query database
    user, err := h.queries.GetUser(r.Context(), id)
    if err != nil {
        h.logger.Errorw("Failed to get user", "id", id, "error", err)
        render.ChiErr(w, r, err, UserNotFound,
            render.WithStatusCode(http.StatusNotFound))
        return
    }
    
    // Convert to response format
    response := UserResponse{
        ID:    user.ID,
        Name:  user.Name,
        Email: user.Email,
    }
    
    render.ChiJSON(w, r, response)
}

// CreateUser creates a new user
func (h *UsersHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    h.logger.Info("Creating user")
    
    // Parse request body
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.Errorw("Failed to decode request", "error", err)
        render.ChiErr(w, r, err, InvalidRequestData,
            render.WithStatusCode(http.StatusBadRequest))
        return
    }
    
    // Create user in database
    user, err := h.queries.CreateUser(r.Context(), db.CreateUserParams{
        Name:  req.Name,
        Email: req.Email,
    })
    if err != nil {
        h.logger.Errorw("Failed to create user", "error", err)
        render.ChiErr(w, r, err, FailedToCreateUser,
            render.WithStatusCode(http.StatusInternalServerError))
        return
    }
    
    // Return created user
    response := UserResponse{
        ID:    user.ID,
        Name:  user.Name,
        Email: user.Email,
    }
    
    render.ChiJSON(w, r, response)
}
```

### Error Codes Pattern

```go
// api/go/_internal/handlers/users/err_codes.go
package users

const (
    InvalidUserID       = "INVALID_USER_ID"
    UserNotFound        = "USER_NOT_FOUND"
    InvalidRequestData  = "INVALID_REQUEST_DATA"
    FailedToCreateUser  = "FAILED_TO_CREATE_USER"
    FailedToUpdateUser  = "FAILED_TO_UPDATE_USER"
    FailedToDeleteUser  = "FAILED_TO_DELETE_USER"
)
```

### Entry Point Pattern

```go
// api/go/entries/users.go
package handler

import (
    "net/http"
    
    "github.com/go-chi/chi/v5"
    appfx "github/huangc28/kikichoice-be/api/go/_internal/fx"
    "github/huangc28/kikichoice-be/api/go/_internal/handlers/users"
    "github/huangc28/kikichoice-be/api/go/_internal/pkg/logger"
    "github/huangc28/kikichoice-be/api/go/_internal/router"
    routerfx "github/huangc28/kikichoice-be/api/go/_internal/router/fx"
    "go.uber.org/fx"
)

// Handle is the main entry point for the users endpoint
func Handle(w http.ResponseWriter, r *http.Request) {
    fx.New(
        logger.TagLogger("users"),
        appfx.CoreConfigOptions,
        routerfx.CoreRouterOptions,
        
        fx.Provide(
            router.AsRoute(users.NewUsersHandler),
        ),
        
        fx.Invoke(func(router *chi.Mux) {
            router.ServeHTTP(w, r)
        }),
    )
}
```

## 🗄️ Database Operations with SQLC

### Adding Database Queries

1. **Define SQL queries** in `api/go/_internal/db/sqlc_queries/queries.sql`:

```sql
-- name: GetUser :one
SELECT id, name, email, created_at, updated_at, deleted_at
FROM users
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: ListUsers :many
SELECT id, name, email, created_at, updated_at
FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateUser :one
INSERT INTO users (name, email)
VALUES ($1, $2)
RETURNING id, name, email, created_at, updated_at, deleted_at;

-- name: UpdateUser :one
UPDATE users
SET name = $2, email = $3, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, name, email, created_at, updated_at, deleted_at;

-- name: DeleteUser :exec
UPDATE users
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;
```

2. **Generate type-safe Go code**:
```bash
make sqlc/generate
```

3. **Use in handlers**:
```go
// Get single user
user, err := h.queries.GetUser(ctx, userID)

// Create user
user, err := h.queries.CreateUser(ctx, db.CreateUserParams{
    Name:  "John Doe",
    Email: "john@example.com",
})

// List users with pagination
users, err := h.queries.ListUsers(ctx, db.ListUsersParams{
    Limit:  10,
    Offset: 0,
})
```

## 🔧 Configuration Management

### Environment Variables
```go
// Access config in handlers
func (h *Handler) SomeMethod() {
    dbURL := h.config.DatabaseURL
    env := h.config.Environment // "development", "production", etc.
}
```

### Common Environment Variables
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `VERCEL_ENV` - Environment detection
- `LOG_LEVEL` - Logging level

## 🛣️ Routing and Middleware

### Adding Global Middleware
```go
// api/go/_internal/router/core.go
func setupRouter() *chi.Mux {
    r := chi.NewRouter()
    
    // Global middlewares
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(60 * time.Second))
    r.Use(cors.Handler(cors.Options{...}))
    
    return r
}
```

### Route Groups
```go
func (h *Handler) RegisterRoutes(r *chi.Mux) {
    r.Route("/api/v1", func(r chi.Router) {
        // Add authentication middleware for this group
        r.Use(authMiddleware)
        
        r.Route("/users", func(r chi.Router) {
            r.Get("/", h.ListUsers)
            r.Post("/", h.CreateUser)
            r.Route("/{id}", func(r chi.Router) {
                r.Get("/", h.GetUser)
                r.Put("/", h.UpdateUser)
                r.Delete("/", h.DeleteUser)
            })
        })
    })
}
```

## 🚀 Deployment with Vercel

### Vercel Configuration
```json
// vercel.json
{
  "functions": {
    "api/go/entries/*.go": {
      "runtime": "vercel-go@3.0.0"
    }
  },
  "rewrites": [
    {
      "source": "/api/users/(.*)",
      "destination": "/api/go/entries/users"
    },
    {
      "source": "/api/health",
      "destination": "/api/go/entries/health"
    }
  ]
}
```

## 🧪 Testing Patterns

### Handler Testing
```go
func TestUsersHandler_GetUser(t *testing.T) {
    // Setup test dependencies
    logger := zap.NewNop().Sugar()
    mockQueries := &MockQueries{}
    
    handler := NewUsersHandler(UsersHandlerParams{
        Logger:  logger,
        Queries: mockQueries,
    })
    
    // Setup test request
    req := httptest.NewRequest("GET", "/users/1", nil)
    req = req.WithContext(chi.NewRouteContext())
    rctx := chi.RouteContext(req.Context())
    rctx.URLParams.Add("id", "1")
    
    w := httptest.NewRecorder()
    
    // Execute
    handler.GetUser(w, req)
    
    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
}
```

## 🔍 Debugging and Logging

### Structured Logging
```go
h.logger.Infow("Processing request",
    "method", r.Method,
    "path", r.URL.Path,
    "user_id", userID)

h.logger.Errorw("Database error",
    "error", err,
    "query", "GetUser",
    "user_id", userID)
```

## 📋 Development Checklist

When creating new endpoints:

- [ ] Create handler package in `api/go/_internal/handlers/`
- [ ] Implement `NewHandler` constructor with `fx.In` params
- [ ] Implement `RegisterRoutes` method
- [ ] Add error codes in `err_codes.go`
- [ ] Create entry point in `api/go/entries/`
- [ ] Add SQL queries to `queries.sql` if needed
- [ ] Run `make sqlc/generate` if queries added
- [ ] Add Vercel routing in `vercel.json`
- [ ] Test locally with `make start/vercel`
- [ ] Write unit tests
- [ ] Update API documentation

## 🚨 Common Pitfalls to Avoid

1. **Don't edit SQLC generated files** (`models.go`, `queries.sql.go`, `db.go`)
2. **Always use dependency injection** - don't create dependencies inside handlers
3. **Use structured logging** - include relevant context
4. **Handle errors properly** - use appropriate HTTP status codes
5. **Validate input data** - never trust user input
6. **Use transactions** for multi-step database operations
7. **Set proper timeouts** for external API calls
8. **Don't forget CORS** for browser requests

---

This guide ensures AI assistants understand the full context and can help developers build consistent, maintainable APIs following the established patterns.
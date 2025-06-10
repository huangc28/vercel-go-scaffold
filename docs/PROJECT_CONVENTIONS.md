# Project Conventions & Standards

This document outlines the coding conventions and standards for the Go Chi Vercel starter project.

## 🏗️ Architecture Principles

### Layered Architecture
```
Entry Points → Handlers → Database → Models
```

### Dependency Injection
- Use Uber FX for all dependency injection
- Constructor pattern: `NewHandler(params HandlerParams)`
- All dependencies via `fx.In` struct tags

## 📁 File Organization

### Handler Structure
```
handlers/
└── feature/
    ├── handler.go      # Main logic
    ├── err_codes.go    # Error constants
    ├── types.go        # Request/response types (optional)
    └── handler_test.go # Tests
```

### Naming Conventions

#### Files
- `handler.go` - Handler implementation
- `err_codes.go` - Error constants
- `types.go` - Request/response structs
- `*_test.go` - Test files

#### Functions & Methods
- Handlers: `Handle`, `HandleCreate`, `HandleUpdate`
- Constructors: `NewHandler`
- Route registration: `RegisterRoutes`

#### Constants
- Error codes: `UPPER_SNAKE_CASE`
- Config keys: `UPPER_SNAKE_CASE`

#### Variables
- Go standard: `camelCase`
- Acronyms: `userID`, `httpClient`, `urlPath`

## 🔧 Handler Implementation

### Required Methods
Every handler must implement:
```go
// Constructor with dependency injection
func NewHandler(params HandlerParams) *Handler

// Route registration
func (h *Handler) RegisterRoutes(r *chi.Mux)

// HTTP handlers
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request)
```

### Error Handling Pattern
```go
// Success response
render.ChiJSON(w, r, data)

// Error response
render.ChiErr(w, r, err, ErrorCode,
    render.WithStatusCode(http.StatusBadRequest))
```

### Request Processing Pattern
```go
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
    // 1. Log request
    h.logger.Info("Processing request")

    // 2. Extract parameters
    id := chi.URLParam(r, "id")

    // 3. Parse request body (if needed)
    var req RequestType
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        render.ChiErr(w, r, err, InvalidRequestData,
            render.WithStatusCode(http.StatusBadRequest))
        return
    }

    // 4. Business logic
    result, err := h.someService.DoSomething(r.Context(), req)
    if err != nil {
        h.logger.Errorw("Operation failed", "error", err)
        render.ChiErr(w, r, err, OperationFailed,
            render.WithStatusCode(http.StatusInternalServerError))
        return
    }

    // 5. Return response
    render.ChiJSON(w, r, result)
}
```

## 🗄️ Database Conventions

### SQLC Query Naming
- Get single: `GetUser`, `GetOrder`
- List multiple: `ListUsers`, `ListOrders`
- Create: `CreateUser`, `CreateOrder`
- Update: `UpdateUser`, `UpdateOrder`
- Delete: `DeleteUser`, `DeleteOrder`

### Query Parameters
```sql
-- Single result
-- name: GetUser :one

-- Multiple results
-- name: ListUsers :many

-- Execute only (no return)
-- name: DeleteUser :exec

-- Return single row after insert/update
-- name: CreateUser :one
```

### Soft Deletes
Use `deleted_at` timestamp:
```sql
WHERE deleted_at IS NULL
```

## 🛣️ Routing Conventions

### Entry Points
- File name matches endpoint: `users.go` → `/api/users`
- One entry point per major resource

### Route Groups
```go
r.Route("/users", func(r chi.Router) {
    r.Get("/", h.ListUsers)        // GET /users
    r.Post("/", h.CreateUser)      // POST /users
    r.Route("/{id}", func(r chi.Router) {
        r.Get("/", h.GetUser)      // GET /users/{id}
        r.Put("/", h.UpdateUser)   // PUT /users/{id}
        r.Delete("/", h.DeleteUser) // DELETE /users/{id}
    })
})
```

### HTTP Methods
- `GET` - Retrieve data
- `POST` - Create new resource
- `PUT` - Update entire resource
- `PATCH` - Update partial resource
- `DELETE` - Remove resource

## 📊 Response Standards

### Success Response Format
```json
{
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com"
  },
  "errors": null
}
```

### Error Response Format
```json
{
  "data": null,
  "errors": [
    {
      "code": "USER_NOT_FOUND",
      "message": "User with ID 123 not found"
    }
  ]
}
```

### HTTP Status Codes
- `200` - Success
- `201` - Created
- `400` - Bad Request (client error)
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `422` - Unprocessable Entity
- `500` - Internal Server Error

## 🔐 Security Best Practices

### Input Validation
- Always validate request data
- Use struct tags for validation rules
- Sanitize user input

### Error Messages
- Don't expose internal details
- Use generic error codes
- Log detailed errors internally

### Database Queries
- Use parameterized queries (SQLC handles this)
- Validate input before querying
- Use transactions for multi-step operations

## 🧪 Testing Conventions

### Test File Structure
```go
func TestHandler_Method(t *testing.T) {
    // Setup
    handler := setupTestHandler()

    // Execute
    req := httptest.NewRequest("GET", "/path", nil)
    w := httptest.NewRecorder()
    handler.Method(w, req)

    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
}
```

### Mock Dependencies
```go
type MockService struct {
    // Mock implementation
}

func (m *MockService) Method() error {
    return nil
}
```

## 📝 Documentation Standards

### Code Comments
```go
// UserHandler handles user-related HTTP requests
type UserHandler struct {
    logger *zap.SugaredLogger
}

// GetUser retrieves a user by ID from the database
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

### API Documentation
Document endpoints with:
- Purpose
- Request format
- Response format
- Error codes
- Examples

## 🚀 Deployment Considerations

### Environment Variables
- Required: `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- Optional: `LOG_LEVEL`, `PORT`

### Vercel Configuration
- Function timeout: 60 seconds max
- Memory limit: Consider function size
- Cold start optimization

---

Following these conventions ensures consistent, maintainable code that AI assistants can easily understand and extend.

# Go Chi Vercel Starter 🚀

A production-ready starter template for building serverless APIs with **Go**, **Chi router**, **Vercel**, **Supabase**, **SQLC**, and **SQLX**.

Perfect for quickly bootstrapping new projects with a robust, scalable architecture.

## 🏗️ Stack

- **Go** - Backend language
- **Chi** - Lightweight HTTP router
- **Vercel** - Serverless deployment platform
- **Supabase** - Database and authentication
- **SQLC** - Type-safe SQL code generation
- **SQLX** - SQL extensions for Go
- **Uber FX** - Dependency injection framework

## 🚀 Quick Start

### Using GitHub Template (Recommended)

1. Click "Use this template" button on GitHub
2. Clone your new repository
3. Run the setup script:

```bash
cd your-new-project
make setup
```

4. Follow the prompts to configure your project

### Manual Setup

1. Clone this repository:
```bash
git clone git@github.com:huangc28/vercel-go-scaffold.git your-project-name
cd your-project-name
```

2. Run the setup script:
```bash
./setup.sh
```

## 📋 Prerequisites

Before you begin, ensure you have:

- **Go 1.21+** installed
- **Node.js 18+** (for Vercel CLI)
- **Vercel CLI** installed: `npm i -g vercel`
- **SQLC** installed: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
- A **Supabase** account and project

## ⚙️ Configuration

1. Copy the environment template:
```bash
cp .env.example .env
```

2. Update `.env` with your configuration:
```bash
# Database Configuration
DB_HOST=your-supabase-host
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your-password
DB_NAME=postgres

```

## 🛠️ Development

### Available Commands

```bash
# Generate type-safe SQL code
make sqlc/generate

# Start Vercel development server
make start/vercel

# Run Go tests
make test

# Lint and format code
make vet

# Deploy to Vercel (preview)
make deploy/vercel/preview

# Deploy to Vercel (production)
make deploy/vercel/prod

# Dump database schema
make supabase/db-dump
```

### Project Structure

```
├── api/
│   └── go/
│       ├── entries/          # Vercel function entry points
│       └── _internal/        # Internal packages
│           ├── configs/      # Configuration management
│           ├── db/          # Database layer (SQLC generated)
│           ├── handlers/    # HTTP handlers
│           ├── middlewares/ # HTTP middlewares
│           ├── pkg/         # Shared utilities
│           └── router/      # Router configuration
├── supabase/
│   ├── migrations/          # Database migrations
│   └── schemas/            # Database schema
├── setup.sh                # Project setup script
├── Makefile                # Development commands
├── sqlc.yaml               # SQLC configuration
└── vercel.json             # Vercel deployment config
```

### Adding New Endpoints

1. Create a handler in `api/go/_internal/handlers/`:
```go
package myhandler

import (
    "net/http"
    "github.com/go-chi/chi/v5"
)

func NewMyHandler() chi.Router {
    r := chi.NewRouter()
    r.Get("/", handleGet)
    return r
}

func handleGet(w http.ResponseWriter, r *http.Request) {
    // Your logic here
}
```

2. Create an entry point in `api/go/entries/`:
```go
package handler

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    // ... your imports
)

func Handle(w http.ResponseWriter, r *http.Request) {
    fx.New(
        // ... dependency injection setup
    )
}
```

3. Update `vercel.json` to add URL mapping:
```json
{
  "rewrites": [
    {
      "source": "/api/my-endpoint",
      "destination": "/api/go/entries/my-endpoint"
    }
  ]
}
```

## 🗄️ Database

### Migrations

Create new migrations using Supabase CLI:
```bash
supabase migration new create_users_table
```

### SQLC Code Generation

1. Write your SQL queries in `api/go/_internal/db/sqlc_queries/queries.sql`
2. Run `make sqlc/generate` to generate type-safe Go code
3. Use the generated code in your handlers

Example query:
```sql
-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (name, email) VALUES ($1, $2) RETURNING *;
```

## 🚢 Deployment

### Preview Deployment
```bash
make deploy/vercel/preview
```

### Production Deployment
```bash
make deploy/vercel/prod
```

### Environment Variables

Set your environment variables in Vercel dashboard:
- Database connection details
- Supabase keys
- Any third-party API keys

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./api/go/_internal/handlers/...
```

## 📚 Architecture

This starter uses a layered architecture with dependency injection:

- **Entry Points** (`api/go/entries/`): Vercel function handlers
- **Handlers** (`_internal/handlers/`): Business logic
- **Repository** (`_internal/db/`): Database operations
- **Middlewares** (`_internal/middlewares/`): Cross-cutting concerns
- **Configuration** (`_internal/configs/`): Environment and app config

### Dependency Injection

Uses Uber FX for clean dependency management:
```go
fx.New(
    logger.TagLogger("my-handler"),
    appfx.CoreConfigOptions,
    routerfx.CoreRouterOptions,
    fx.Provide(
        router.AsRoute(myhandler.NewMyHandler),
    ),
    fx.Invoke(func(router *chi.Mux) {
        router.ServeHTTP(w, r)
    }),
)
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Useful Links

- [Go Documentation](https://golang.org/doc/)
- [Chi Router](https://github.com/go-chi/chi)
- [Vercel Go Runtime](https://vercel.com/docs/functions/serverless-functions/runtimes/go)
- [Supabase Documentation](https://supabase.com/docs)
- [SQLC Documentation](https://docs.sqlc.dev/)

## 🤖 AI Assistant Integration

This project includes comprehensive documentation for AI assistants (GitHub Copilot, Cursor, etc.) to understand the project structure and conventions:

- **`.cursorrules`** - Primary configuration file that Cursor automatically loads for context
- **`docs/AI_DEVELOPMENT_GUIDE.md`** - Detailed development guide with examples and patterns
- **`.vscode/settings.json`** - VS Code configuration optimized for Go development and Copilot

These files contain all the context AI assistants need to help you build APIs following the project's conventions without requiring lengthy explanations each time.

---

**Happy coding!** 🎉
# AI Assistant Quick Reference

This project includes AI-friendly documentation to help assistants understand the codebase structure and conventions.

## Key Files for AI Context

1. **`.cursorrules`** (Primary) - Cursor's automatic context file
2. **`docs/AI_DEVELOPMENT_GUIDE.md`** - Comprehensive development guide
3. **`.vscode/settings.json`** - VS Code configuration for optimal development

## Quick Commands

- `make sqlc/generate` - Generate type-safe database code
- `make start/vercel` - Start development server
- `make test` - Run tests
- `make deploy/vercel/preview` - Deploy preview

## Project Structure Summary

```
api/go/
├── entries/        # Vercel serverless function endpoints
└── _internal/      # Core application code
    ├── handlers/   # Business logic (your main work area)
    ├── db/        # Database layer (SQLC generated + manual)
    ├── configs/   # Configuration management
    ├── middlewares/ # HTTP middlewares
    ├── pkg/       # Shared utilities
    └── router/    # Router configuration
```

## Adding New Endpoints

1. Create handler in `api/go/_internal/handlers/yourhandler/`
2. Create entry point in `api/go/entries/your-endpoint.go`
3. Add Vercel routing to `vercel.json`
4. Add SQL queries to `queries.sql` if needed
5. Run `make sqlc/generate` to update database code

For detailed instructions, see the full documentation files above.

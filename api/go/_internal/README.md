# Go Code Structure in WebVitals Edge Functions

This document explains the structure and purpose of Go code in this repository, specifically focusing on the `go_internal` directory and its integration with Vercel Edge Functions.

## Overview

This project implements business logic for CDC (Change Data Capture) handlers using Go. The `go_internal` directory contains the core Go code that processes CDC requests within the Vercel serverless environment.

## Project Structure

The repository follows the standard Go project layout as recommended by [golang-standards/project-layout](https://github.com/golang-standards/project-layout), with one key modification:

- The standard `internal` directory has been renamed to `go_internal` to work around Go's import restrictions.

### Why `go_internal` Instead of `internal`?

In Go, the `internal` directory has special meaning. As explained in the golang-standards documentation:

> You can optionally add a bit of extra structure to your internal packages to separate your shared and non-shared internal code... You use internal directories to make packages private. If you put a package inside an internal directory, then other packages can't import it unless they share a common ancestor. And it's the only directory named in Go's documentation and has special compiler treatment.

Due to the nature of the Vercel deployment environment and our project requirements, we needed to make these "internal" packages accessible from various parts of our codebase, which would be restricted if we used the standard `internal` naming. Renaming to `go_internal` allows us to maintain the conceptual organization while bypassing Go's import restrictions.

## Go Functions in Vercel

### Vercel Configuration

The `vercel.json` file contains configuration for our Go functions:

```json
"functions": {
  "api/**/*.go": {
    "memory": 1024,
    "maxDuration": 10
  }
}
```

This allocates 1024MB of memory to each Go function and sets a maximum execution duration of 10 seconds.

The following rewrites are configured to map friendly URLs to Go function endpoints:

```json
"rewrites": [
  { "source": "/hello", "destination": "/api/go/hello/core" },
  { "source": "/v1/sites", "destination": "/go/api/sites/core" }
]
```

## Directory Structure of `go_internal`

The `go_internal` directory contains several packages:

- `router/`: Handles HTTP routing and request dispatching
- `middlewares/`: Contains middleware functions for request processing
- `pkg/`: Shared utilities and helper functions
- `handlers/`: Business logic implementations for CDC handlers
- `db/`: Database access and operations
- `fx/`: Dependency injection configuration using Uber's fx library
- `configs/`: Application configuration

## Getting Started with Go Code

To work with the Go code in this project:

1. Understand that the entry points are in `api/go/` directory
2. Core implementation logic is in the `go_internal/` directory
3. The project uses [Uber's fx](https://github.com/uber-go/fx) for dependency injection
4. Changes to Go code will be deployed as Vercel serverless functions

## Adding New Functions

When adding new Go functions:

1. Create a new handler in the appropriate subdirectory of `go_internal/handlers/`
2. Register the handler with the router in a new or existing entry point file
3. Update `vercel.json` if you need to add new URL mappings

For more detailed information about Go modules, refer to the official Go documentation.

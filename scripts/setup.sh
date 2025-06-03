#!/bin/bash

set -e

echo "🚀 Go-Chi-Vercel Starter Setup"
echo "================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    print_error "This doesn't appear to be a git repository. Please run this script from the root of your cloned repository."
    exit 1
fi

# Get current directory name as default project name
DEFAULT_PROJECT_NAME=$(basename "$PWD")

# Prompt for project details
echo ""
echo "Please provide the following information for your new project:"
echo ""

read -p "Enter your module name (e.g., github.com/username/project-name): " MODULE_NAME
if [ -z "$MODULE_NAME" ]; then
    print_error "Module name is required"
    exit 1
fi

read -p "Enter your project name (default: $DEFAULT_PROJECT_NAME): " PROJECT_NAME
PROJECT_NAME=${PROJECT_NAME:-$DEFAULT_PROJECT_NAME}

echo ""
print_status "Setting up project with:"
print_status "  Module name: $MODULE_NAME"
print_status "  Project name: $PROJECT_NAME"
echo ""

# Backup original files
print_status "Creating backup of original files..."
cp go.mod go.mod.bak

# Update go.mod
print_status "Updating go.mod..."
sed -i.bak "s|module github.com/huangc28/vercel-go-scaffold|module $MODULE_NAME|g" go.mod

# Find and replace all Go import references
print_status "Updating import statements in Go files..."

# Replace the main module references
find . -name "*.go" -type f -exec sed -i.bak "s|github\.com/huangc28/vercel-go-scaffold|$MODULE_NAME|g" {} \;

# Replace the inconsistent webvitals references
find . -name "*.go" -type f -exec sed -i.bak "s|github\.com/webvitals-sh/webvitals-edge-funcs|$MODULE_NAME|g" {} \;

# Update any project-specific names in comments or strings
find . -name "*.go" -type f -exec sed -i.bak "s|vercel-go-scaffold|$PROJECT_NAME|g" {} \;
find . -name "*.md" -type f -exec sed -i.bak "s|vercel-go-scaffold|$PROJECT_NAME|g" {} \;

# Clean up backup files
print_status "Cleaning up backup files..."
find . -name "*.bak" -type f -delete

# Update go dependencies
print_status "Updating Go dependencies..."
go mod tidy

# Create .env.example if it doesn't exist
if [ ! -f ".env.example" ]; then
    print_status "Creating .env.example file..."
    cat > .env.example << 'EOF'
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=myapp

# Supabase Configuration
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key

# Vercel Environment
VERCEL_ENV=development

# AWS Configuration (optional)
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
AWS_S3_BUCKET_REGION=
AWS_S3_SNAPSHOT_BUCKET=

# Auth Configuration (optional)
CLERK_SECRET_KEY=

# External Services (optional)
STARBURST_HOST=
STARBURST_PORT=
STARBURST_CATALOG=
STARBURST_SCHEMA=
STARBURST_USER=
STARBURST_PASSWORD=

INNGEST_EVENT_KEY=
INNGEST_APP_ID=
EOF
fi

# Create .gitignore if it doesn't exist or update it
if [ ! -f ".gitignore" ]; then
    print_status "Creating .gitignore file..."
    cat > .gitignore << 'EOF'
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with `go test -c`
*.test

# Output of the go coverage tool, specifically when used with LiteIDE
*.out

# Dependency directories (remove the comment below to include it)
# vendor/

# Go workspace file
go.work

# Environment variables
.env
.env.local

# Vercel
.vercel

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Logs
*.log

# Database
*.db
*.sqlite
*.sqlite3

# Backup files
*.bak
EOF
else
    print_status ".gitignore already exists, skipping..."
fi

# Update README if it exists
if [ -f "README.md" ]; then
    print_status "Updating README.md..."
    sed -i.bak "s|go-chi-vercel-starter|$PROJECT_NAME|g" README.md
    rm -f README.md.bak
fi

print_status "Project setup completed successfully! 🎉"
echo ""
echo "Next steps:"
echo "1. Copy .env.example to .env and fill in your configuration"
echo "2. Set up your Supabase project and update the connection details"
echo "3. Run 'make sqlc/generate' to generate database code"
echo "4. Run 'make start/vercel' to start the development server"
echo "5. Deploy to Vercel with 'make deploy/vercel/preview'"
echo ""
print_warning "Don't forget to update your git remote origin if this is a new repository:"
print_warning "  git remote set-url origin $MODULE_NAME.git"
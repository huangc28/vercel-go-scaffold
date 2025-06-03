# Template Usage Guide

This repository is designed to be used as a GitHub template for quickly bootstrapping new Go + Chi + Vercel + Supabase projects.

## 🚀 Quick Start

### Option 1: Using GitHub Template (Recommended)

1. **Click "Use this template"** button on the GitHub repository page
2. **Create your new repository** with your desired name
3. **Clone your new repository**:
   ```bash
   git clone https://github.com/yourusername/your-new-project.git
   cd your-new-project
   ```
4. **Run the setup script**:
   ```bash
   ./setup.sh
   ```
5. **Follow the prompts** to configure your project

### Option 2: Manual Clone

1. **Clone this repository**:
   ```bash
   git clone https://github.com/huangc28/go-chi-vercel-starter.git your-project-name
   cd your-project-name
   ```
2. **Run the setup script**:
   ```bash
   ./setup.sh
   ```
3. **Update git remote** (if needed):
   ```bash
   git remote set-url origin https://github.com/yourusername/your-new-project.git
   ```

## 🛠️ What the Setup Script Does

The `setup.sh` script automatically:

1. **Updates module name** in `go.mod` and all import statements
2. **Fixes inconsistent import paths** throughout the codebase
3. **Updates project references** in documentation
4. **Creates `.env.example`** with all required environment variables
5. **Runs `go mod tidy`** to update dependencies
6. **Cleans up backup files** created during the process

## 📋 After Setup

Once the setup script completes, you should:

1. **Copy environment file**:
   ```bash
   cp .env.example .env
   ```

2. **Configure your environment variables** in `.env`:
   - Database connection details
   - Supabase project information
   - Any third-party service keys

3. **Set up your Supabase project**:
   - Create a new project at [supabase.com](https://supabase.com)
   - Update the database connection details in `.env`

4. **Start development**:
   ```bash
   make start/vercel
   ```

## 🔧 Available Commands

After setup, you can use these Make commands:

```bash
make help                    # Show all available commands
make setup                   # Run the setup script
make sqlc/generate          # Generate type-safe SQL code
make start/vercel           # Start development server
make test                   # Run tests
make test/coverage          # Run tests with coverage
make vet                    # Run go vet
make deploy/vercel/preview  # Deploy to Vercel (preview)
make deploy/vercel/prod     # Deploy to Vercel (production)
```

## 📁 Project Structure

After setup, your project will have this structure:

```
your-project/
├── api/go/
│   ├── entries/           # Vercel function entry points
│   └── _internal/         # Internal packages
├── supabase/
│   ├── migrations/        # Database migrations
│   └── schemas/          # Database schema
├── .github/              # GitHub templates
├── setup.sh              # Setup script (can be deleted after use)
├── .env.example          # Environment variables template
├── Makefile              # Development commands
└── README.md             # Project documentation
```

## 🗑️ Cleanup

After successful setup, you can optionally:

1. **Delete the setup script** (no longer needed):
   ```bash
   rm setup.sh
   ```

2. **Delete this template usage guide**:
   ```bash
   rm TEMPLATE_USAGE.md
   ```

3. **Commit your changes**:
   ```bash
   git add .
   git commit -m "Initial project setup"
   git push origin main
   ```

## 🤝 Contributing Back

If you make improvements to the template itself, consider contributing back:

1. Fork the original template repository
2. Make your improvements
3. Submit a pull request

## 📚 Next Steps

- Read the main `README.md` for detailed development instructions
- Set up your database schema in `supabase/migrations/`
- Create your first API endpoint in `api/go/_internal/handlers/`
- Deploy to Vercel with `make deploy/vercel/preview`

Happy coding! 🎉
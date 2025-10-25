# dailyprog

A command-line tool for quickly scaffolding new programming projects with pre-configured templates.

## Features

- 🚀 **Multi-Language Support** - Go, Python, and Rust templates included
- 📝 **Multiple Templates** - Different project types per language (basic, webserver, flask, etc.)
- ⚙️ **Configurable** - Customize templates and user information via JSON or command-line flags
- � **Embedded Templates** - All templates embedded in the binary - works anywhere
- 🎨 **Template Generation** - Generate customizable template directories
- �🔧 **Auto-Setup** - Runs initialization commands (go mod init, pip install, cargo build)
- 💻 **VS Code Integration** - Automatically opens projects in VS Code
- 📅 **Date-Based Organization** - Projects organized by date with version control
- 👤 **Smart Defaults** - Uses current username when no config specified

## Quick Start

```bash
# Build the tool
go build -o dailyprog main.go

# List available templates
./dailyprog --list

# Create a Go project (uses current username as author)
./dailyprog myproject

# Create with custom author
./dailyprog --author "Alice" myproject

# Create a Python Flask web app
./dailyprog --lang python --template flask mywebapp

# Create a Rust project
./dailyprog --lang rust myrust
```

## Installation

### From Source

1. Clone this repository
2. Build the binary:

   ```bash
   go build -o dailyprog main.go
   ```

3. Optionally move to your PATH:

   ```bash
   sudo mv dailyprog /usr/local/bin/
   ```

4. (Optional) Generate and customize templates:

   ```bash
   dailyprog --generate-template ./my-templates
   # Edit files in ./my-templates/
   dailyprog --templates ./my-templates/templates.json --user-config ./my-templates/user-config.json myproject
   ```

**Note:** Templates are embedded in the binary, so the `_buildin` directory is not required for the compiled program to work.

### Using go install

- Install [Go](https://go.dev/dl/)
- Execute: `go install github.com/524D/dailyprog@latest`
- The `dailyprog` executable will be in `~/go/bin/dailyprog`

## Available Templates

### Go

- **basic** - Simple Hello World program
- **webserver** - HTTP web server using net/http

### Python

- **basic** - Simple Python script
- **flask** - Flask web application with virtual environment

### Rust

- **basic** - Simple Rust program with Cargo

## What It Does

When you run `dailyprog`, it:

1. Creates a directory named `~/dailyprog/YYYYMMDD-projectname` (with version numbers if needed)
2. Copies and processes template files from embedded filesystem
3. Substitutes template variables with your information (author, copyright, etc.)
4. Runs initialization commands (go mod init, pip install, cargo build, etc.)
5. Opens the new project in Visual Studio Code

## Configuration

### Default Behavior (No Configuration Needed)

By default, `dailyprog` uses:

- **Author**: Current logged-in username (from `$USER`)
- **Copyright**: "Copyright (c) [YEAR] [USERNAME]. All rights reserved."
- **Templates**: Embedded templates (Go, Python, Rust)

### Command-Line Overrides

Override settings on a per-project basis:

```bash
# Override author only
dailyprog --author "Alice Smith" myproject

# Override both author and copyright
dailyprog --author "Bob Jones" --copyright "Copyright 2025 Acme Corp" myproject
```

### Custom Configuration Files

Generate customizable configuration and templates:

```bash
# Generate template directory
dailyprog --generate-template ./my-templates

# This creates:
# ./my-templates/
#   ├── templates.json      # Template definitions
#   ├── user-config.json    # User information
#   └── templates/          # Template files
#       ├── go/
#       ├── python/
#       └── rust/

# Edit the files as needed, then use:
dailyprog --templates ./my-templates/templates.json \
          --user-config ./my-templates/user-config.json \
          myproject
```

### User Configuration Format

Edit `user-config.json` with your personal information:

```json
{
  "author": "Your Name",
  "copyright": "Copyright (c) 2025 Your Name. All rights reserved.",
  "email": "your.email@example.com",
  "organization": "Your Organization"
}
```

**Priority Order** (highest to lowest):

1. Command-line flags (`--author`, `--copyright`)
2. Custom user-config file (`--user-config`)
3. Current username (default when no config specified)

### Templates Configuration

The `templates.json` file defines available languages and templates. See [IMPLEMENTATION.md](IMPLEMENTATION.md) for details on adding custom templates.

## Command-Line Options

```text
Usage: dailyprog [options] [project-name]

Options:
  -l, --lang string          Programming language (default: go)
  -T, --template string      Template to use (default: basic)
  -d, --dir string          Base directory (default: ~/dailyprog)
  -t, --templates string    Templates config file (uses embedded if not specified)
  -u, --user-config string  User config file (uses embedded if not specified)
  -g, --generate-template   Generate template directory at specified path
      --author string       Override author name
      --copyright string    Override copyright text
  -v, --verbose             Show detailed output
  -V, --version             Print version
      --list                List available languages and templates
```

## Examples

```bash
# Create project with default settings (uses current username)
./dailyprog myproject

# Create multiple projects
./dailyprog project1 project2 project3

# Create with custom author and copyright
./dailyprog --author "Alice Smith" --copyright "MIT License" myapp

# Create with custom directory
./dailyprog --dir ~/myprojects myapp

# Create Python Flask app with verbose output
./dailyprog -v --lang python --template flask myflaskapp

# Create Go web server
./dailyprog --lang go --template webserver api-server

# Generate customizable templates
./dailyprog --generate-template ./custom-templates

# Use custom templates
./dailyprog --templates ./custom-templates/templates.json \
            --user-config ./custom-templates/user-config.json \
            myproject
```

## Documentation

- **[USAGE.md](USAGE.md)** - Detailed usage guide
- **[IMPLEMENTATION.md](IMPLEMENTATION.md)** - Technical implementation details
- **[QUICKREF.md](QUICKREF.md)** - Quick reference guide

## Requirements

- Go 1.16+ (for building)
- VS Code with `code` command in PATH
- Language-specific tools for chosen templates:
  - Go: `go` command
  - Python: `python3`, `pip`
  - Rust: `cargo`

## Contributing

### Adding a New Template

1. Generate a template directory:

   ```bash
   ./dailyprog --generate-template ./my-templates
   ```

2. Add your template files in `./my-templates/templates/<language>/<template>/`:

   ```bash
   mkdir -p ./my-templates/templates/go/mytemplate
   # Add your template files here
   ```

3. Edit `./my-templates/templates.json` to define the new template:

   ```json
   {
     "languages": {
       "go": {
         "templates": {
           "mytemplate": {
             "name": "My Custom Template",
             "description": "Description of my template",
             "files": [
               {
                 "source": "go/mytemplate/main.go",
                 "dest": "main.go"
               }
             ],
             "postCreateSteps": []
           }
         }
       }
     }
   }
   ```

4. Use Go template syntax (`{{.Variable}}`) for variable substitution:
   - `{{.ProgName}}` - Project name
   - `{{.Author}}` - Author name
   - `{{.Copyright}}` - Copyright text
   - `{{.Email}}` - Email address
   - `{{.Organization}}` - Organization name
   - `{{.Date}}` - Current date (YYYY-MM-DD)

5. Test your template:

   ```bash
   ./dailyprog --templates ./my-templates/templates.json \
               --user-config ./my-templates/user-config.json \
               --lang go --template mytemplate test
   ```

See [IMPLEMENTATION.md](IMPLEMENTATION.md) for detailed instructions.

## Credits

Uses the following Go packages:

- github.com/spf13/pflag - Command-line flag parsing

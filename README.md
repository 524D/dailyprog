# dailyprog

A command-line tool for quickly scaffolding new programming projects with pre-configured templates.

## Features

- 🚀 **Multi-Language Support** - Go, Python, and Rust templates included
- 📝 **Multiple Templates** - Different project types per language (basic, webserver, flask, etc.)
- ⚙️ **Configurable** - Customize templates and user information via JSON
- 🔧 **Auto-Setup** - Runs initialization commands (go mod init, pip install, cargo build)
- 💻 **VS Code Integration** - Automatically opens projects in VS Code
- 📅 **Date-Based Organization** - Projects organized by date with version control

## Quick Start

```bash
# Build the tool
go build -o dailyprog main.go

# List available templates
./dailyprog --list

# Create a Go project (default)
./dailyprog myproject

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

4. Configure your user information:

   ```bash
   nano _buildin/user-config.json
   ```

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
2. Copies and processes template files with your personal information
3. Runs initialization commands (go mod init, pip install, cargo build, etc.)
4. Opens the new project in Visual Studio Code

## Configuration

### User Configuration (`_buildin/user-config.json`)

Edit this file with your personal information:

```json
{
  "author": "Your Name",
  "copyright": "Copyright (c) 2025 Your Name. All rights reserved.",
  "email": "your.email@example.com",
  "organization": "Your Organization"
}
```

This information will be automatically inserted into your template files.

### Templates Configuration (`_buildin/templates.json`)

Defines available languages and templates. See [IMPLEMENTATION.md](IMPLEMENTATION.md) for details on adding custom templates.

## Command-Line Options

```text
Usage: dailyprog [options] [project-name]

Options:
  -l, --lang string         Programming language (default: go)
  -T, --template string     Template to use (default: basic)
  -d, --dir string         Base directory (default: ~/dailyprog)
  -t, --templates string   Templates config file (default: _buildin/templates.json)
  -u, --user-config string User config file (default: _buildin/user-config.json)
  -v, --verbose            Show detailed output
  -V, --version            Print version
      --list               List available languages and templates
```

## Examples

```bash
# Create multiple projects
./dailyprog project1 project2 project3

# Create with custom directory
./dailyprog --dir ~/myprojects myapp

# Create Python Flask app with verbose output
./dailyprog -v --lang python --template flask myflaskapp

# Create Go web server
./dailyprog --lang go --template webserver api-server
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

To add a new template:

1. Edit `_buildin/templates.json`
2. Create template files in `_buildin/templates/<language>/<template>/`
3. Use Go template syntax (`{{.Variable}}`) for variable substitution
4. Add post-create steps if needed
5. Test with `./dailyprog --lang <language> --template <template> test`

See [IMPLEMENTATION.md](IMPLEMENTATION.md) for detailed instructions.

## Credits

Uses the following Go packages:

- github.com/spf13/pflag - Command-line flag parsing

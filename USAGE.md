# Daily Programming Tool - Usage Guide

## Overview

This tool quickly scaffolds new programming projects with pre-configured templates for different languages and project types. It creates a timestamped directory, populates it with template files, runs initialization commands, and opens the project in VS Code.

## Quick Start

### Basic Usage

```bash
# Create a Go project with default (basic) template
./dailyprog

# Create a Python Flask web app
./dailyprog --lang python --template flask

# Create a Go web server with a custom name
./dailyprog --lang go --template webserver mywebapp

# List all available languages and templates
./dailyprog --list
```

## Command-Line Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--list` | | false | List all available languages and templates |
| `--lang` | `-l` | `go` | Programming language (go, python, rust) |
| `--template` | `-T` | `basic` | Template to use (basic, webserver, flask) |
| `--dir` | `-d` | `~/dailyprog` | Base directory for new projects |
| `--templates` | `-t` | `_buildin/templates.json` | Path to templates configuration file |
| `--user-config` | `-u` | `_buildin/user-config.json` | Path to user configuration file |
| `--verbose` | `-v` | false | Show detailed output |
| `--version` | `-V` | false | Print version and exit |

## Available Templates

### Go
- **basic**: Simple Hello World program
  - Creates `main.go` with basic program structure
  - Initializes Go module with `go mod init` and `go mod tidy`
  - Includes VS Code launch configuration

- **webserver**: HTTP web server
  - Creates a basic HTTP server using `net/http`
  - Listens on port 8080
  - Includes VS Code launch configuration

### Python
- **basic**: Simple Python script
  - Creates `main.py` with basic structure
  - Includes VS Code launch configuration

- **flask**: Flask web application
  - Creates `app.py` with Flask routes
  - Includes `requirements.txt` with Flask dependency
  - Creates Python virtual environment
  - Installs dependencies automatically
  - Includes VS Code launch configuration for Flask debugging

### Rust
- **basic**: Simple Rust program
  - Creates `src/main.rs` with basic program
  - Generates `Cargo.toml` with project metadata
  - Runs `cargo build` to compile
  - Includes VS Code launch configuration

## Configuration Files

### User Configuration (`_buildin/user-config.json`)

Customize these values with your information:

```json
{
  "author": "Your Name",
  "copyright": "Copyright (c) 2025 Your Name. All rights reserved.",
  "email": "your.email@example.com",
  "organization": "Your Organization"
}
```

These values are substituted into template files using Go template syntax:
- `{{.Author}}` - Author name
- `{{.Copyright}}` - Copyright message
- `{{.Email}}` - Email address
- `{{.Organization}}` - Organization name
- `{{.ProgName}}` - Project name (auto-generated)
- `{{.Date}}` - Current date (auto-generated)

### Templates Configuration (`_buildin/templates.json`)

Defines available languages and templates. Each template specifies:
- **files**: Template files to copy and process
- **postCreateSteps**: Commands to run after file creation
  - `remove`: Delete a file
  - `exec`: Run a command

## Examples

### Create Multiple Projects

```bash
# Creates: 20251024-project1, 20251024-project2
./dailyprog project1 project2 project3
```

### Create with Custom Directory

```bash
./dailyprog --dir ~/myprojects myapp
```

### Create Python Flask App with Verbose Output

```bash
./dailyprog --lang python --template flask --verbose myflaskapp
```

### Create Go Web Server

```bash
./dailyprog --lang go --template webserver myserver
```

## Directory Structure

Projects are created with this pattern:
- If no name provided: `dailyprog-YYYYMMDD`
- If name provided: `YYYYMMDD-name`
- If directory exists, appends `-1`, `-2`, etc.

## Adding Custom Templates

1. Edit `_buildin/templates.json` to add new language or template
2. Create template files in `_buildin/templates/<language>/<template>/`
3. Use Go template syntax for variable substitution
4. Define post-create steps for initialization commands

Example template file:
```go
// {{.Copyright}}
// Author: {{.Author}}

package main

func main() {
    // Your code here
}
```

## Troubleshooting

### Template file not found
- Ensure template files exist in `_buildin/templates/`
- Check that paths in `templates.json` match actual file locations

### Post-create commands fail
- Verify required tools are installed (go, python3, cargo, etc.)
- Use `--verbose` flag to see detailed command output

### VS Code doesn't open
- Ensure `code` command is in your PATH
- Test with: `code --version`

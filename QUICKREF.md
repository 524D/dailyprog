# Quick Reference

## List Available Templates
```bash
./dailyprog --list
```

## Create Projects

### Go
```bash
# Basic Hello World
./dailyprog myproject

# Web Server
./dailyprog --lang go --template webserver myserver
```

### Python
```bash
# Basic Script
./dailyprog --lang python --template basic myscript

# Flask Web App
./dailyprog --lang python --template flask myflaskapp
```

### Rust
```bash
# Basic Program
./dailyprog --lang rust --template basic myrust
```

## Configuration Files

Edit these files to customize your projects:

- **`_buildin/user-config.json`** - Your name, email, copyright
- **`_buildin/templates.json`** - Add/modify languages and templates

## Template Variables

Use these in your template files:
- `{{.Author}}` - Your name
- `{{.Copyright}}` - Copyright message  
- `{{.Email}}` - Email address
- `{{.Organization}}` - Organization name
- `{{.ProgName}}` - Project name
- `{{.Date}}` - Current date

## Flags

| Flag | Description |
|------|-------------|
| `--list` | Show all templates |
| `--lang <lang>` | Choose language (go/python/rust) |
| `--template <name>` | Choose template (basic/webserver/flask) |
| `--verbose` | Show detailed output |
| `--dir <path>` | Change base directory |

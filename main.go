package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	flag "github.com/spf13/pflag"
)

//go:embed _buildin
//go:embed _buildin/templates/*/*/.vscode/*
var embeddedFS embed.FS

// This program executes the fixed actions for starting a "program of the day"
// * create a directory named prog_of_the_day/yyyymmdd
// * create the standard files based on templates for different programming languages
// * supports multiple templates per language (basic, webserver, etc.)
// * Open new folder in Visual Studio Code

const maxVers = 1000 // Maximum number of version directories to create

// TemplateData holds the data to be substituted in templates
type TemplateData struct {
	ProgName     string
	Date         string
	Author       string
	Copyright    string
	Email        string
	Organization string
}

// UserConfig holds user-specific configuration
type UserConfig struct {
	Author       string `json:"author"`
	Copyright    string `json:"copyright"`
	Email        string `json:"email"`
	Organization string `json:"organization"`
}

// TemplateFile represents a file to be copied from template
type TemplateFile struct {
	Source string `json:"source"`
	Dest   string `json:"dest"`
}

// PostCreateStep represents an action to run after creating files
type PostCreateStep struct {
	Type    string   `json:"type"`    // "exec" or "remove"
	Command []string `json:"command"` // for exec type
	Path    string   `json:"path"`    // for remove type
}

// Template represents a project template
type Template struct {
	Name            string           `json:"name"`
	Description     string           `json:"description"`
	Files           []TemplateFile   `json:"files"`
	PostCreateSteps []PostCreateStep `json:"postCreateSteps"`
}

// Language represents a programming language configuration
type Language struct {
	Name          string              `json:"name"`
	FileExtension string              `json:"fileExtension"`
	Templates     map[string]Template `json:"templates"`
}

// TemplatesConfig holds all templates configuration
type TemplatesConfig struct {
	Languages map[string]Language `json:"languages"`
}
type options struct {
	verbose          bool
	version          bool
	dir              string
	templatesFile    string
	userConfigFile   string
	language         string
	templateName     string
	list             bool
	generateTemplate string
	author           string
	copyright        string
}

var opt options

func printUsage() {
	fmt.Println("dailyprog - Quickly scaffold new programming projects with pre-configured templates")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  dailyprog [OPTIONS] [PROJECT_NAME...]")
	fmt.Println()
	fmt.Println("DESCRIPTION:")
	fmt.Println("  Creates a new project directory with template files for your chosen language.")
	fmt.Println("  Projects are organized by date (YYYYMMDD-projectname) in ~/dailyprog/")
	fmt.Println("  Automatically opens the project in VS Code when complete.")
	fmt.Println()
	fmt.Println("OPTIONS:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Create a Go project with default settings")
	fmt.Println("  dailyprog myproject")
	fmt.Println()
	fmt.Println("  # Create with custom author")
	fmt.Println("  dailyprog --author \"Alice Smith\" myproject")
	fmt.Println()
	fmt.Println("  # Create a Python Flask web app")
	fmt.Println("  dailyprog --lang python --template flask mywebapp")
	fmt.Println()
	fmt.Println("  # Create a Rust project")
	fmt.Println("  dailyprog --lang rust myrust")
	fmt.Println()
	fmt.Println("  # List all available templates")
	fmt.Println("  dailyprog --list")
	fmt.Println()
	fmt.Println("  # Generate customizable template directory")
	fmt.Println("  dailyprog --generate-template ./my-templates")
	fmt.Println()
	fmt.Println("  # Use custom templates")
	fmt.Println("  dailyprog --templates ./my-templates/templates.json \\")
	fmt.Println("            --user-config ./my-templates/user-config.json myproject")
	fmt.Println()
	fmt.Println("AVAILABLE LANGUAGES:")
	fmt.Println("  go      - Go programming language (templates: basic, webserver)")
	fmt.Println("  python  - Python (templates: basic, flask)")
	fmt.Println("  rust    - Rust (templates: basic)")
	fmt.Println()
	fmt.Println("For more information, visit: https://github.com/524D/dailyprog")
}

func init() {
	flag.Usage = printUsage
	flag.BoolVarP(&opt.verbose, "verbose", "v", false, "Show what's being done")
	flag.BoolVarP(&opt.version, "version", "V", false, "Print version and exit")
	flag.StringVarP(&opt.dir, "dir", "d", "~/dailyprog", "Base directory name where new program is created.")
	flag.StringVarP(&opt.templatesFile, "templates", "t", "", "Path to templates configuration file (uses embedded if not specified)")
	flag.StringVarP(&opt.userConfigFile, "user-config", "u", "", "Path to user configuration file (uses embedded if not specified)")
	flag.StringVarP(&opt.language, "lang", "l", "go", "Programming language to use (e.g., go, python, rust)")
	flag.StringVarP(&opt.templateName, "template", "T", "basic", "Template to use (e.g., basic, webserver, flask)")
	flag.BoolVar(&opt.list, "list", false, "List available languages and templates")
	flag.StringVarP(&opt.generateTemplate, "generate-template", "g", "", "Generate template directory at specified path (e.g., ./my-templates)")
	flag.StringVar(&opt.author, "author", "", "Override author name from user-config")
	flag.StringVar(&opt.copyright, "copyright", "", "Override copyright from user-config")
}

// readConfigFile reads a config file from filesystem or embedded FS
func readConfigFile(path string) ([]byte, error) {
	// If path is provided and exists on filesystem, use it
	if path != "" {
		if data, err := os.ReadFile(path); err == nil {
			return data, nil
		}
	}

	// Otherwise, use embedded filesystem
	var embeddedPath string
	if path == "" || strings.HasPrefix(path, "_buildin/") {
		embeddedPath = path
	} else {
		// If a path was provided but not found, still try embedded as fallback
		embeddedPath = path
	}

	// Default to embedded paths if none specified
	if embeddedPath == "" {
		return nil, fmt.Errorf("no path specified and no default available")
	}

	data, err := embeddedFS.ReadFile(embeddedPath)
	if err != nil {
		return nil, fmt.Errorf("file not found in filesystem or embedded FS: %w", err)
	}

	return data, nil
}

// readTemplateFile reads a template file from embedded FS or filesystem
func readTemplateFile(relativePath string, templatesBasePath string) ([]byte, error) {
	// If a custom templates path is provided, try to read from filesystem first
	if templatesBasePath != "" {
		fullPath := filepath.Join(templatesBasePath, "templates", relativePath)
		if data, err := os.ReadFile(fullPath); err == nil {
			return data, nil
		}
	}

	// Otherwise, read from embedded FS
	fullPath := filepath.Join("_buildin", "templates", relativePath)
	// Normalize path for embedded FS (use forward slashes)
	fullPath = filepath.ToSlash(fullPath)

	data, err := embeddedFS.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("can't read template file %s: %w", fullPath, err)
	}

	return data, nil
}

// generateTemplateDirectory extracts all embedded templates to a directory
func generateTemplateDirectory(targetPath string) error {
	// Create the target directory
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("can't create target directory: %w", err)
	}

	// Walk through the embedded _buildin directory
	err := fs.WalkDir(embeddedFS, "_buildin", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate the relative path from _buildin
		relPath, err := filepath.Rel("_buildin", path)
		if err != nil {
			return err
		}

		// Skip the root _buildin directory itself
		if relPath == "." {
			return nil
		}

		// Create the target path
		targetFilePath := filepath.Join(targetPath, relPath)

		if d.IsDir() {
			// Create directory
			if err := os.MkdirAll(targetFilePath, 0755); err != nil {
				return fmt.Errorf("can't create directory %s: %w", targetFilePath, err)
			}
			if opt.verbose {
				fmt.Printf("Created directory: %s\n", targetFilePath)
			}
		} else {
			// Read file from embedded FS
			content, err := embeddedFS.ReadFile(path)
			if err != nil {
				return fmt.Errorf("can't read embedded file %s: %w", path, err)
			}

			// Write file to target
			if err := os.WriteFile(targetFilePath, content, 0644); err != nil {
				return fmt.Errorf("can't write file %s: %w", targetFilePath, err)
			}
			if opt.verbose {
				fmt.Printf("Created file: %s\n", targetFilePath)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking embedded filesystem: %w", err)
	}

	return nil
}

func main() {
	flag.Parse()

	// Handle --generate-template flag
	if opt.generateTemplate != "" {
		if opt.verbose {
			fmt.Printf("Generating template directory at: %s\n", opt.generateTemplate)
		}
		if err := generateTemplateDirectory(opt.generateTemplate); err != nil {
			log.Fatalf("Error generating template directory: %v\n", err)
		}
		fmt.Printf("Template directory successfully generated at: %s\n", opt.generateTemplate)
		fmt.Println("\nYou can now:")
		fmt.Printf("  1. Modify the templates in: %s/templates/\n", opt.generateTemplate)
		fmt.Printf("  2. Edit configuration files: %s/templates.json and %s/user-config.json\n", opt.generateTemplate, opt.generateTemplate)
		fmt.Printf("  3. Use them with: dailyprog --templates %s/templates.json --user-config %s/user-config.json\n", opt.generateTemplate, opt.generateTemplate)
		return
	}

	// Use embedded config files if paths not specified
	templatesPath := opt.templatesFile
	if templatesPath == "" {
		templatesPath = "_buildin/templates.json"
	}

	userConfigPath := opt.userConfigFile
	if userConfigPath == "" {
		userConfigPath = "_buildin/user-config.json"
	}

	// Load configurations
	templatesConfig, err := loadTemplatesConfig(templatesPath)
	if err != nil {
		log.Fatalln("Error loading templates config:", err)
	}

	userConfig, err := loadUserConfig(userConfigPath)
	if err != nil {
		log.Fatalln("Error loading user config:", err)
	}

	// Apply command-line overrides or defaults
	if opt.author != "" {
		userConfig.Author = opt.author
	} else if opt.userConfigFile == "" && userConfig.Author == "Your Name" {
		// No user-config file and no --author flag, use current user
		currentUser, err := user.Current()
		if err == nil && currentUser.Username != "" {
			userConfig.Author = currentUser.Username
		}
	}

	if opt.copyright != "" {
		userConfig.Copyright = opt.copyright
	} else if opt.userConfigFile == "" && strings.Contains(userConfig.Copyright, "Your Name") {
		// No user-config file and no --copyright flag, use current user in copyright
		currentUser, err := user.Current()
		if err == nil && currentUser.Username != "" {
			year := time.Now().Year()
			userConfig.Copyright = fmt.Sprintf("Copyright (c) %d %s. All rights reserved.", year, currentUser.Username)
		}
	}

	// Determine the templates base directory (parent directory of templates.json if specified)
	var templatesBasePath string
	if opt.templatesFile != "" {
		templatesBasePath = filepath.Dir(opt.templatesFile)
	}

	// Handle --list flag
	if opt.list {
		listTemplates(templatesConfig)
		return
	}

	// Validate language and template
	lang, ok := templatesConfig.Languages[opt.language]
	if !ok {
		log.Fatalf("Language '%s' not found. Use --list to see available languages.\n", opt.language)
	}

	_, ok = lang.Templates[opt.templateName]
	if !ok {
		log.Fatalf("Template '%s' not found for language '%s'. Use --list to see available templates.\n", opt.templateName, opt.language)
	}

	homeDirNative, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Can't get name of home directory:", err)
	}
	homeDir := filepath.ToSlash(homeDirNative)
	dateStr := time.Now().Format("20060102")

	dailyDir := strings.Replace(opt.dir, "~", homeDir, 1)
	if len(flag.Args()) > 0 {
		for _, a := range flag.Args() {
			dailyprogDir := filepath.Join(dailyDir, dateStr+"-"+a)
			createDailyProg(dailyprogDir, a, templatesConfig, userConfig, templatesBasePath)
		}
	} else {
		dailyprogDir := filepath.Join(dailyDir, "dailyprog-"+dateStr)
		createDailyProg(dailyprogDir, "dailyprog-"+dateStr, templatesConfig, userConfig, templatesBasePath)
	}
}

// loadTemplatesConfig loads the templates configuration from JSON file or embedded FS
func loadTemplatesConfig(path string) (*TemplatesConfig, error) {
	data, err := readConfigFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading templates file: %w", err)
	}

	var config TemplatesConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing templates JSON: %w", err)
	}

	return &config, nil
}

// loadUserConfig loads user configuration from JSON file or embedded FS
func loadUserConfig(path string) (*UserConfig, error) {
	data, err := readConfigFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading user config file: %w", err)
	}

	var config UserConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing user config JSON: %w", err)
	}

	return &config, nil
}

// listTemplates displays all available languages and templates
func listTemplates(config *TemplatesConfig) {
	fmt.Println("Available Languages and Templates:")
	fmt.Println("==================================")

	for langKey, lang := range config.Languages {
		fmt.Printf("\n%s (%s)\n", lang.Name, langKey)
		fmt.Println(strings.Repeat("-", len(lang.Name)+len(langKey)+3))

		for tmplKey, tmpl := range lang.Templates {
			fmt.Printf("  %-15s - %s\n", tmplKey, tmpl.Description)
		}
	}

	fmt.Println("\nUsage: dailyprog --lang <language> --template <template> [name]")
}

func createDailyProg(dailyprogDir string, progName string, templatesConfig *TemplatesConfig, userConfig *UserConfig, templatesBasePath string) error {
	// Check if dir exists, and continue to append a version string until we have a non-existing dir
	vers := 0
	versStr := ""
	for _, err := os.Stat(dailyprogDir + versStr); !os.IsNotExist(err); _, err = os.Stat(dailyprogDir + versStr) {
		if vers > maxVers {
			log.Fatalln("All directories " + dailyprogDir + " to " + dailyprogDir + versStr + " seem to exist, quitting.")
		}
		vers++
		versStr = "-" + strconv.Itoa(vers)
	}
	dailyprogDir = dailyprogDir + versStr

	// Create the main directory
	err := os.MkdirAll(dailyprogDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("can't create directory: %w", err)
	}

	// Get the selected language and template
	lang := templatesConfig.Languages[opt.language]
	tmpl := lang.Templates[opt.templateName]

	// Prepare template data
	templateData := TemplateData{
		ProgName:     progName,
		Date:         time.Now().Format("2006-01-02"),
		Author:       userConfig.Author,
		Copyright:    userConfig.Copyright,
		Email:        userConfig.Email,
		Organization: userConfig.Organization,
	}

	// Process each template file
	for _, tf := range tmpl.Files {
		destPath := filepath.Join(dailyprogDir, tf.Dest)

		// Create destination directory if needed
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
			return fmt.Errorf("can't create directory %s: %w", destDir, err)
		}

		// Read template file from embedded FS or filesystem
		content, err := readTemplateFile(tf.Source, templatesBasePath)
		if err != nil {
			return fmt.Errorf("can't read template file %s: %w", tf.Source, err)
		}

		// Process template
		tmplParsed, err := template.New("file").Parse(string(content))
		if err != nil {
			return fmt.Errorf("can't parse template %s: %w", tf.Source, err)
		}

		var buf bytes.Buffer
		if err := tmplParsed.Execute(&buf, templateData); err != nil {
			return fmt.Errorf("can't execute template %s: %w", tf.Source, err)
		}

		// Write processed file
		if err := os.WriteFile(destPath, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("can't write file %s: %w", destPath, err)
		}

		if opt.verbose {
			fmt.Printf("Created: %s\n", destPath)
		}
	}

	// Change to the new directory
	if err := os.Chdir(dailyprogDir); err != nil {
		return fmt.Errorf("can't change to directory %s: %w", dailyprogDir, err)
	}

	// Execute post-create steps
	for _, step := range tmpl.PostCreateSteps {
		switch step.Type {
		case "remove":
			targetPath := filepath.Join(dailyprogDir, step.Path)
			if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
				log.Printf("Warning: couldn't remove %s: %v\n", targetPath, err)
			}
			if opt.verbose {
				fmt.Printf("Removed: %s\n", targetPath)
			}

		case "exec":
			// Process command arguments through template engine
			var processedCmd []string
			for _, arg := range step.Command {
				tmplParsed, err := template.New("cmd").Parse(arg)
				if err != nil {
					return fmt.Errorf("can't parse command template: %w", err)
				}
				var buf bytes.Buffer
				if err := tmplParsed.Execute(&buf, templateData); err != nil {
					return fmt.Errorf("can't execute command template: %w", err)
				}
				processedCmd = append(processedCmd, buf.String())
			}

			if opt.verbose {
				fmt.Printf("Executing: %s\n", strings.Join(processedCmd, " "))
			}

			cmd := exec.Command(processedCmd[0], processedCmd[1:]...)
			cmd.Stdout = io.Discard
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("command failed: %s: %w", strings.Join(processedCmd, " "), err)
			}
		}
	}

	// Determine the main file to open based on the first file in template
	var mainFile string
	if len(tmpl.Files) > 0 {
		mainFile = filepath.Join(dailyprogDir, tmpl.Files[0].Dest)
	}

	// Open VS Code
	cmdArgs := []string{"--disable-workspace-trust", "-n", dailyprogDir}
	if mainFile != "" {
		cmdArgs = append(cmdArgs, mainFile)
	}

	if opt.verbose {
		fmt.Printf("Opening VS Code: code %s\n", strings.Join(cmdArgs, " "))
	}

	cmd := exec.Command("code", cmdArgs...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("can't open VS Code: %w", err)
	}

	fmt.Printf("Created %s project in: %s\n", lang.Name, dailyprogDir)
	return nil
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	flag "github.com/spf13/pflag"
)

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

// Golang code for program that is created (kept for backward compatibility)
const mainTemplate = `package main

import (
	"fmt"
)
	
func main() {
	fmt.Println("Hello world!")
}	
`

// VS Code launch file (kept for backward compatibility)
const launchJSON = `{
    // Use IntelliSense to learn about possible attributes.
    // Hover to view descriptions of existing attributes.
    // For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Package",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${fileDirname}"
        }
    ]
}
`

type options struct {
	verbose        bool
	version        bool
	dir            string
	templatesFile  string
	userConfigFile string
	language       string
	templateName   string
	list           bool
}

var opt options

func init() {
	flag.BoolVarP(&opt.verbose, "verbose", "v", false, "Show what's being done")
	flag.BoolVarP(&opt.version, "version", "V", false, "Print version and exit")
	flag.StringVarP(&opt.dir, "dir", "d", "~/dailyprog", "Base directory name where new program is created.")
	flag.StringVarP(&opt.templatesFile, "templates", "t", "_buildin/templates.json", "Path to templates configuration file")
	flag.StringVarP(&opt.userConfigFile, "user-config", "u", "_buildin/user-config.json", "Path to user configuration file")
	flag.StringVarP(&opt.language, "lang", "l", "go", "Programming language to use (e.g., go, python, rust)")
	flag.StringVarP(&opt.templateName, "template", "T", "basic", "Template to use (e.g., basic, webserver, flask)")
	flag.BoolVar(&opt.list, "list", false, "List available languages and templates")
}

func main() {
	flag.Parse()

	// Get executable directory to resolve relative paths
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalln("Can't get executable path:", err)
	}
	exeDir := filepath.Dir(exePath)

	// Resolve template and config file paths
	templatesPath := resolveConfigPath(opt.templatesFile, exeDir)
	userConfigPath := resolveConfigPath(opt.userConfigFile, exeDir)

	// Load configurations
	templatesConfig, err := loadTemplatesConfig(templatesPath)
	if err != nil {
		log.Fatalln("Error loading templates config:", err)
	}

	userConfig, err := loadUserConfig(userConfigPath)
	if err != nil {
		log.Fatalln("Error loading user config:", err)
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
			createDailyProg(dailyprogDir, a, templatesConfig, userConfig, exeDir)
		}
	} else {
		dailyprogDir := filepath.Join(dailyDir, "dailyprog-"+dateStr)
		createDailyProg(dailyprogDir, "dailyprog-"+dateStr, templatesConfig, userConfig, exeDir)
	}
}

// resolveConfigPath resolves a config file path relative to executable if needed
func resolveConfigPath(configPath, exeDir string) string {
	if filepath.IsAbs(configPath) {
		return configPath
	}
	// Try relative to current directory first
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}
	// Try relative to executable directory
	return filepath.Join(exeDir, configPath)
}

// loadTemplatesConfig loads the templates configuration from JSON file
func loadTemplatesConfig(path string) (*TemplatesConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading templates file: %w", err)
	}

	var config TemplatesConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing templates JSON: %w", err)
	}

	return &config, nil
}

// loadUserConfig loads user configuration from JSON file
func loadUserConfig(path string) (*UserConfig, error) {
	data, err := os.ReadFile(path)
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

func createDailyProg(dailyprogDir string, progName string, templatesConfig *TemplatesConfig, userConfig *UserConfig, exeDir string) error {
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
	templatesBaseDir := filepath.Join(exeDir, "_buildin", "templates")
	for _, tf := range tmpl.Files {
		sourcePath := filepath.Join(templatesBaseDir, tf.Source)
		destPath := filepath.Join(dailyprogDir, tf.Dest)

		// Create destination directory if needed
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
			return fmt.Errorf("can't create directory %s: %w", destDir, err)
		}

		// Read template file
		content, err := os.ReadFile(sourcePath)
		if err != nil {
			return fmt.Errorf("can't read template file %s: %w", sourcePath, err)
		}

		// Process template
		tmplParsed, err := template.New("file").Parse(string(content))
		if err != nil {
			return fmt.Errorf("can't parse template %s: %w", sourcePath, err)
		}

		var buf bytes.Buffer
		if err := tmplParsed.Execute(&buf, templateData); err != nil {
			return fmt.Errorf("can't execute template %s: %w", sourcePath, err)
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

package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	flag "github.com/spf13/pflag"
)

// This program executes the fixed actions for starting a "program of the day"
// * create a directory named prog_of_the_day/yyyymmdd
// * create the standard files to get going (for go: main.go, go.mod, go.sum, vscode debug launch fiule)
// * Open new folder in Visual Studio Code

// Golang code for program that is created
const mainTemplate = `package main

import (
	"fmt"
)
	
func main() {
	fmt.Println("Hello world!")
}	
`

// VS Code launch file
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

const maxVers = 1000 // Maximum number of version directories to create

type options struct {
	verbose bool
	version bool
	dir     string
}

var opt options

func init() {
	flag.BoolVarP(&opt.verbose, "verbose", "v", false, "Show what's being done")
	flag.BoolVarP(&opt.version, "version", "V", false, "Print version and exit")
	flag.StringVarP(&opt.dir, "dir", "d", "~/dailyprog", "Base directory name where new program is created.")
}

func main() {
	flag.Parse()
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
			createDailyProg(dailyprogDir, a)
		}
	} else {
		dailyprogDir := filepath.Join(dailyDir, "dailyprog-"+dateStr)
		createDailyProg(dailyprogDir, "dailyprog-"+dateStr)
	}
}

func createDailyProg(dailyprogDir string, progName string) error {
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
	vsCodeDir := filepath.Join(dailyprogDir, ".vscode")
	err := os.MkdirAll(vsCodeDir, os.ModePerm) // Also creates dailyprogDir
	if err != nil {
		log.Fatalln("Can't create directory:", err)
	}
	dailyprogDirMain := filepath.Join(dailyprogDir, "main.go")
	f, err := os.Create(dailyprogDirMain)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	_, err = f.WriteString(mainTemplate)

	if err != nil {
		log.Fatal(err)
	}

	vsCodeLaunch := filepath.Join(vsCodeDir, "launch.json")
	f2, err := os.Create(vsCodeLaunch)

	if err != nil {
		log.Fatal(err)
	}

	defer f2.Close()

	_, err = f2.WriteString(launchJSON)

	if err != nil {
		log.Fatal(err)
	}

	os.Chdir(dailyprogDir)

	goModFn := filepath.Join(dailyprogDir, "go.mod")
	_ = os.Remove(goModFn)

	modName := "dummy/" + progName
	// Execute go mod init dummy/dailyprog-yyyymmdd
	_, err = exec.Command("go", "mod", "init", modName).Output()
	if err != nil {
		log.Fatal(err)
	}
	// Execute go mod tidy
	_, err = exec.Command("go", "mod", "tidy").Output()
	if err != nil {
		log.Fatal(err)
	}
	// Open VS code
	_, err = exec.Command("code", "--disable-workspace-trust", "-n", dailyprogDir, dailyprogDirMain).Output()
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

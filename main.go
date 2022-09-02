package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

// This program executes the fixed actions for starting a "program of teh day"
// * create a directory named prog_of_the_day/yyyymmdd
// * create the standard files to get going (for go: main.go, go.mod, go.sum, vscode debug launch fiule)
// * Open new folder in Visual Studio Code

const baseDir = "dailyprog"
const maxVers = 1000 // Maximum number of verion directories to create

func main() {
	mainTemplate := `package main

import (
	"fmt"
)
	
func main() {
	fmt.Println("Hello world!")
}	
`
	launchJSON := `{
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
	homeDirNative, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Can't get name of home directory:", err)
	}
	homeDir := filepath.ToSlash(homeDirNative)
	dateDir := time.Now().Format("20060102")
	dailyprogDir := filepath.Join(homeDir, baseDir, dateDir)
	// Check if dir existst, and continue to append a version string until we have a non-existing dir
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
	err = os.MkdirAll(vsCodeDir, os.ModePerm) // Also creates dailyprogDir
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

	modName := "dummy/dailyprog-" + dateDir
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
}

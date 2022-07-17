package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// This program executes the fixed actions for starting a "program of teh day"
// * create a directory named prog_of_the_day/yyyymmdd
// * create the standard files to get going (for go: main.go, go.mod, go.sum, vscode debug launch fiule)
// * Open new folder in Visual Studio Code

const baseDir = "dailyprog"

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
	_, err = exec.Command("code", "-n", dailyprogDir).Output()
	if err != nil {
		log.Fatal(err)
	}
}

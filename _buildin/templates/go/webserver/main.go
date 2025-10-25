package main

// {{.Copyright}}
// Author: {{.Author}}

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/pflag"
)

// openBrowser opens the default web browser to the specified URL
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default:
		return fmt.Errorf("unsupported platform")
	}

	return exec.Command(cmd, args...).Start()
}

// handleRoot handles requests to the root path
func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!\n")
	fmt.Fprintf(w, "Path: %s\n", r.URL.Path)
}

func main() {
	port := pflag.IntP("port", "p", 8080, "Port to listen on")
	startBrowser := pflag.Bool("startbrowser", true, "Open the default web browser")
	pflag.Parse()

	http.HandleFunc("/", handleRoot)

	addr := fmt.Sprintf(":%d", *port)
	url := fmt.Sprintf("http://localhost%s", addr)
	fmt.Printf("Starting server on %s\n", url)

	// Start browser after a short delay to ensure server is ready
	if *startBrowser {
		go func() {
			time.Sleep(100 * time.Millisecond)
			if err := openBrowser(url); err != nil {
				log.Printf("Failed to open browser: %v\n", err)
			}
		}()
	}

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

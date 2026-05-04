package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

func main() {
	// Args
	args := os.Args[1:]
	if len(args) > 0 {
		checkPath(args[0])
	} else {
		fmt.Println("\nNo arguments provided.")
	}

	for {
		clearScreen()

		fmt.Printf("=== System Monitor (Last update: %s) ===\n\n", time.Now().Format(time.RFC3339))

		// Hostname
		hostname, err := os.Hostname()
		if err != nil {
			fmt.Printf("Error getting hostname: %s%v%s\n", colorRed, err, colorReset)
			return
		}

		fmt.Printf("Hostname: \t%s\n", hostname)

		// Network check
		checkNetwork("https://google.com")

		// User ID and username
		fmt.Printf("User ID: \t%d\n", os.Getuid())

		time.Sleep(5 * time.Second)
	}

}

func clearScreen() {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error clearing screen: %s%v%s\n", colorRed, err, colorReset)
	}
}

func checkNetwork(url string) {
	// Create client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Make request
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Network: \t%s[DOWN]%s (Error: %v)\n", colorRed, colorReset, err)
		return
	}

	defer resp.Body.Close()

	fmt.Printf("Network: \t%s[UP]%s (Status: %s)\n", colorGreen, colorReset, resp.Status)
}

func checkPath(path string) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		fmt.Printf("Path '%s': \t%s[NOT FOUND]%s\n", path, colorRed, colorReset)
	} else {
		fmt.Printf("Path '%s': \t%s[EXISTS]%s\n", path, colorGreen, colorReset)
	}
}

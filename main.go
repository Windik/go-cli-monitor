package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	fmt.Println("=== System Monitor ===")

	// Hostname
	hostname, err := os.Hostname()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Hostname: [NOT FOUND]")
			return
		} else {
			fmt.Printf("Error getting hostname: %v\n", err)
		}

		return
	}

	fmt.Printf("Hostname: \t%s\n", hostname)

	// Network check
	checkNetwork("https://google.com")

	// User ID and username
	fmt.Printf("User ID: \t%d\n", os.Getuid())

	// Args
	args := os.Args[1:]
	if len(args) > 0 {
		checkPath(args[0])
	} else {
		fmt.Println("\nNo arguments provided.")
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
		fmt.Printf("Network [DOWN] (Error: %v)\n", err)
		return
	}

	defer resp.Body.Close()

	fmt.Printf("Network: [UP] (Status: %s)\n", resp.Status)
}

func checkPath(path string) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		fmt.Printf("Path '%s': [NOT FOUND]\n", path)
	} else {
		fmt.Printf("Path '%s': [EXISTS]\n", path)
	}
}

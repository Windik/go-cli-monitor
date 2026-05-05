package main

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"go-cli-monitor/internal/config"

	"github.com/getlantern/systray"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

type TargetMenuItem struct {
	URL  string
	Item *systray.MenuItem
}

//go:embed green_circle_icon_32.png
var iconGreen []byte

//go:embed red_circle_icon_32.png
var iconRed []byte

//go:embed orange_circle_icon_32.png
var iconOrange []byte

var cfg *config.Config

func main() {
	var err error
	cfg, err = config.LoadConfig()

	if err != nil {
		fmt.Printf("Error loading config: %s%v%s\n", colorRed, err, colorReset)
		return
	}

	fmt.Printf("Config loaded: %v\n", cfg)
	fmt.Printf("Check interval: %d\n", cfg.CheckInterval)
	fmt.Printf("Network monitoring targets: %v\n", cfg.Targets)

	// Args
	if args := os.Args[1:]; len(args) > 0 {
		checkPath(args[0])
	} else {
		fmt.Println("\nNo arguments provided.")
	}

	systray.Run(onReady, onExit)
}

func onReady() {
	if len(iconGreen) == 0 || len(iconRed) == 0 {
		fmt.Println("Error: icon files not embed.")
		systray.Quit()
		return
	}

	systray.SetIcon(iconGreen)
	systray.SetTitle(cfg.DefaultTitle)

	// Hostname
	hostname, err := os.Hostname()
	uid := os.Getuid()

	if err != nil {
		fmt.Printf("Error getting hostname: %s%v%s\n", colorRed, err, colorReset)
		return
	}

	// Tray tooltip
	systray.SetTooltip(fmt.Sprintf("Host: %s | User ID: %d", hostname, uid))

	// Menu items with system information and disabled state
	systray.AddMenuItem(fmt.Sprintf("💻 Host: %s", hostname), "Your PC name").Disable()
	systray.AddMenuItem(fmt.Sprintf("👤 User ID: %d", uid), "ID текущего пользователя").Disable()

	systray.AddSeparator()

	networkMenu := systray.AddMenuItem("🌐 Network", "Network status")

	var menuItems []TargetMenuItem

	for _, target := range cfg.Targets {
		subItem := networkMenu.AddSubMenuItem(fmt.Sprintf("⏳ %s", target), target)
		subItem.Disable() // Делаем его просто информационным

		menuItems = append(menuItems, TargetMenuItem{
			URL:  target,
			Item: subItem,
		})
	}

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("❌ Quit", "Close application")

	go func() {
		for {
			clearScreen()

			fmt.Printf("=== System Monitor (Last update: %s) ===\n\n", time.Now().Format(time.RFC3339))

			// Hostname
			fmt.Printf("Hostname: \t%s\n", hostname)
			// User ID and username
			fmt.Printf("User ID: \t%d\n", os.Getuid())

			upCount := 0
			targetsCount := len(cfg.Targets)

			for _, targetItem := range menuItems {
				isUp := checkNetworkAndReturn(targetItem.URL)

				if isUp {
					upCount++
					// Обновляем текст в подменю (зеленый кружок-эмодзи)
					targetItem.Item.SetTitle(fmt.Sprintf("🟢 %s", targetItem.URL))
				} else {
					// Обновляем текст в подменю (красный кружок-эмодзи)
					targetItem.Item.SetTitle(fmt.Sprintf("🔴 %s", targetItem.URL))
				}
			}

			if upCount == targetsCount {
				systray.SetIcon(iconGreen)
				systray.SetTooltip(fmt.Sprintf("All %d targets are UP | Host: %s", targetsCount, hostname))
			} else if upCount < targetsCount {
				systray.SetIcon(iconOrange)
				systray.SetTooltip(fmt.Sprintf("Warning: %d/%d targets UP | Host: %s", upCount, targetsCount, hostname))
			} else {
				systray.SetIcon(iconRed)
				systray.SetTooltip(fmt.Sprintf("Critical: All targets are DOWN! | Host: %s", hostname))
			}

			time.Sleep(time.Duration(cfg.CheckInterval) * time.Second)
		}
	}()

	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

func onExit() {
	fmt.Print("Applications succesfuly closed")
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

func checkNetworkAndReturn(url string) bool {
	// Create client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Make request
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Connection for url - %s is \t%s[DOWN]%s (Error: %v)\n", url, colorRed, colorReset, err)
		return false
	}

	defer resp.Body.Close()

	fmt.Printf("Connection for url - %s is \t%s[UP]%s (Status: %s)\n", url, colorGreen, colorReset, resp.Status)

	return resp.StatusCode == http.StatusOK
}

func checkPath(path string) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Path '%s': \t%s[NOT FOUND]%s\n", path, colorRed, colorReset)
		} else {
			fmt.Printf("Path '%s': \t%s[EXISTS]%s\n", path, colorGreen, colorReset)
		}
		return
	}

	fmt.Printf("Path '%s': \t%s[EXISTS]%s\n", path, colorGreen, colorReset)
}

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
	"go-cli-monitor/internal/logger"

	"github.com/getlantern/systray"
)

const version = "1.0.0"

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

type CheckResult struct {
	URL  string
	IsUp bool
}

type SystemInfo struct {
	Hostname string
	UID      int
	OS       string
	Arch     string
}

type AppInfo struct {
	StartTime time.Time
	Version   string
}

type NetworkStats struct {
	Target      string
	TotalChecks int
	UpCount     int
	DownCount   int
	LastCheck   time.Time
}

type Reporter interface {
	Report() string
}

func (a AppInfo) Uptime() string {
	duration := time.Since(a.StartTime).Round(time.Second)
	return duration.String()
}

func (a AppInfo) String() string {
	return fmt.Sprintf("%s | Started: %s | Uptime: %s", a.Version, a.StartTime.Format("15:04:05"), a.Uptime())
}

func (s SystemInfo) Summary() string {
	return fmt.Sprintf("Hostname: %s, UID: %d, OS: %s, Arch: %s", s.Hostname, s.UID, s.OS, s.Arch)
}

func (c CheckResult) StatusLabel() string {
	if c.IsUp {
		return "🟢 " + c.URL
	}
	return "🔴 " + c.URL
}

func (a AppInfo) Report() string {
	return fmt.Sprintf("[APP] %s | Uptime: %s", a.Version, a.Uptime())
}

func (s SystemInfo) Report() string {
	return fmt.Sprintf("[SYS] %s", s.Summary())
}

func (n *NetworkStats) RecordCheck(isUp bool) {
	n.TotalChecks++
	if isUp {
		n.UpCount++
	} else {
		n.DownCount++
	}
	n.LastCheck = time.Now()
}

func (n NetworkStats) SuccessRate() float64 {
	if n.TotalChecks == 0 {
		return 0
	}

	return float64(n.UpCount) / float64(n.TotalChecks) * 100
}

func (n NetworkStats) LastCheckAgo() string {
	if n.LastCheck.IsZero() {
		return "never"
	}

	return fmt.Sprintf("%s", time.Since(n.LastCheck).Round(time.Second))
}

// Print Slice of reporters
func printAllReports(reporters []Reporter) {
	fmt.Println("=== Startup Report ===")

	for _, r := range reporters {
		fmt.Printf(">> %s\n", r.Report())
	}

	fmt.Println("==================")
}

func (n NetworkStats) Report() string {
	return fmt.Sprintf("[NET] \t%s | Success Rate: %.2f%% | Last Check: %s",
		n.Target, n.SuccessRate(), n.LastCheckAgo())
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

		errorMessage := fmt.Sprintf("Error loading config: %s", err)
		logger.Log(logger.LevelError, errorMessage)

		return
	}

	logger.Log(logger.LevelInfo, "Config loaded")

	fmt.Printf("Log: \t\t%s\n", logger.GetLogPath())

	fmt.Printf("Config loaded: %v\n", cfg)
	fmt.Printf("Check interval: %d\n", cfg.CheckInterval)
	fmt.Printf("Network monitoring targets: %v\n", cfg.Targets)

	// Args
	if args := os.Args[1:]; len(args) > 0 {
		checkPath(args[0])
	} else {
		logger.Log(logger.LevelInfo, "No arguments provided.")
		fmt.Println("\nNo arguments provided.")
	}

	app := AppInfo{
		StartTime: time.Now(),
		Version:   version,
	}

	info, err := getSystemInfo()

	if err != nil {
		fmt.Printf("Error getting system info: %s%v%s\n", colorRed, err, colorReset)
		logger.Log(logger.LevelError, fmt.Sprintf("Error getting system info: %s", err))
		return
	}

	reporters := []Reporter{app, info}

	for _, target := range cfg.Targets {
		reporters = append(reporters, NetworkStats{Target: target})
	}

	printAllReports(reporters)

	systray.Run(func() {
		onReady(app)
	}, onExit)
}

func onReady(app AppInfo) {
	if len(iconGreen) == 0 || len(iconRed) == 0 {
		errorMessage := "Error: icon files not embed."

		fmt.Println(errorMessage)
		logger.Log(logger.LevelError, errorMessage)

		systray.Quit()
		return
	}

	systray.SetIcon(iconGreen)
	systray.SetTitle(cfg.DefaultTitle)

	// System Info
	info, err := getSystemInfo()
	stats := make([]NetworkStats, len(cfg.Targets))

	for i, target := range cfg.Targets {
		stats[i] = NetworkStats{Target: target}
	}

	if err != nil {
		fmt.Printf("Error getting hostname: %s%v%s\n", colorRed, err, colorReset)

		errorMessage := fmt.Sprintf("Error getting hostname: %s", err)
		logger.Log(logger.LevelError, errorMessage)

		return
	}

	// Tray tooltip
	systray.SetTooltip(fmt.Sprintf("Host: %s | User ID: %d", info.Hostname, info.UID))

	// Menu items with system information and disabled state
	systray.AddMenuItem(fmt.Sprintf("💻 Host: %s", info.Hostname), "Your PC name").Disable()
	systray.AddMenuItem(fmt.Sprintf("👤 User ID: %d", info.UID), "Current User ID").Disable()
	systray.AddMenuItem(fmt.Sprintf("🖥️ OS: %s | Arch: %s", info.OS, info.Arch), "Operating System and Architecture").Disable()

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

			fmt.Printf("App: \t\t%s\n", app)

			fmt.Printf("System: \t%s\n", info.Summary())

			upCount := 0
			targetsCount := len(cfg.Targets)

			resultsChan := make(chan CheckResult, targetsCount)

			for _, targetItem := range menuItems {
				go func(url string) {
					isUp := checkNetworkAndReturn(url)

					resultsChan <- CheckResult{URL: url, IsUp: isUp}
				}(targetItem.URL)
			}

			for i := 0; i < targetsCount; i++ {
				res := <-resultsChan // Read one result from channel

				for _, targetItem := range menuItems {
					if targetItem.URL == res.URL {
						if res.IsUp {
							upCount++
						}
						targetItem.Item.SetTitle(res.StatusLabel())
					}
				}

				for j := range stats {
					if stats[j].Target == res.URL {
						stats[j].RecordCheck(res.IsUp)
					}
				}
			}

			if upCount == targetsCount {
				systray.SetIcon(iconGreen)
				systray.SetTooltip(fmt.Sprintf("All %d targets are UP | Host: %s", targetsCount, info.Hostname))
			} else if upCount < targetsCount {
				systray.SetIcon(iconOrange)
				systray.SetTooltip(fmt.Sprintf("Warning: %d/%d targets UP | Host: %s", upCount, targetsCount, info.Hostname))
			} else {
				systray.SetIcon(iconRed)
				systray.SetTooltip(fmt.Sprintf("Critical: All targets are DOWN! | Host: %s", info.Hostname))
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
	logger.Log(logger.LevelInfo, "Applications succesfuly closed")
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

		errorMessage := fmt.Sprintf("Error clearing screen: %s", err)
		logger.Log(logger.LevelError, errorMessage)
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

		errorMessage := fmt.Sprintf("Connection for url - %s [DOWN] - [ERROR] %v", url, err)
		logger.Log(logger.LevelError, errorMessage)

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
			logger.Log(logger.LevelError, fmt.Sprintf("Path '%s' not found: %v", path, err))
		} else {
			fmt.Printf("Path '%s': \t%s[ACCESS ERROR]%s\n", path, colorRed, colorReset)
			logger.Log(logger.LevelError, fmt.Sprintf("Path '%s' access error: %v", path, err))
		}

		return
	}

	fmt.Printf("Path '%s': \t%s[EXISTS]%s\n", path, colorGreen, colorReset)
}

func getSystemInfo() (SystemInfo, error) {
	hostname, err := os.Hostname()

	if err != nil {
		return SystemInfo{}, err
	}

	uid := os.Getuid()

	return SystemInfo{
		Hostname: hostname,
		UID:      uid,
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
	}, nil
}

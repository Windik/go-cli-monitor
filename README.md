# go-cli-monitor

A lightweight, cross-platform system monitor written in Go that lives in your system tray and provides real-time network status and basic system information directly in the terminal.

---

## Features

- 🟢 **System Tray Integration** — Displays a green/red circle icon in the system tray reflecting live network status.
- 🌐 **Network Monitoring** — Periodically checks internet connectivity by pinging `https://google.com` every 5 seconds.
- 💻 **System Info** — Shows the machine hostname and current user ID in both the terminal and the tray tooltip/menu.
- 📁 **Path Checker** — Accepts an optional CLI argument to verify whether a given file or directory path exists on the filesystem.
- 🎨 **Colored Terminal Output** — Uses ANSI color codes for clear, readable status output (`[UP]` in green, `[DOWN]` in red, etc.).
- 🖥️ **Cross-Platform** — Supports macOS, Linux, and Windows (screen clearing uses `clear` or `cls` accordingly).
- 📦 **Self-Contained Binary** — Tray icons are embedded directly into the binary at compile time using Go's `//go:embed` directive.

---

## Requirements

- **Go** 1.21+ (module path declares `go 1.26.1`)
- A desktop environment with system tray support (required by `getlantern/systray`)

---

## Installation

```/dev/null/sh#L1-4
# Clone the repository
git clone https://github.com/your-username/go-cli-monitor.git
cd go-cli-monitor
```

### Build

```/dev/null/sh#L1-2
go build -o go-cli-monitor .
```

### Run

```/dev/null/sh#L1-5
# Run without arguments (network monitor + tray only)
./go-cli-monitor

# Run with a path argument to check existence
./go-cli-monitor /path/to/check
```

---

## Usage

### Basic — System Monitor

Running the binary without arguments starts the monitor in both the terminal and the system tray:

```/dev/null/text#L1-8
=== System Monitor (Last update: 2024-06-01T12:00:00Z) ===

Hostname:       my-macbook.local
User ID:        501
Network:        [UP] (Status: 200 OK)
```

The terminal refreshes every **5 seconds**. The system tray icon turns:
- 🟢 **Green** — internet is reachable
- 🔴 **Red** — internet is unreachable

### Path Argument

Pass a filesystem path as the first argument to check whether it exists:

```/dev/null/sh#L1-4
./go-cli-monitor /etc/hosts
# Path '/etc/hosts':    [EXISTS]

./go-cli-monitor /nonexistent/file
# Path '/nonexistent/file':    [NOT FOUND]
```

### Quitting

Click the **❌ Quit** item in the system tray menu to gracefully exit the application.

---

## System Tray Menu

| Item | Description |
|------|-------------|
| 💻 Host: `<hostname>` | Displays the machine's hostname (disabled/info only) |
| 👤 User ID: `<uid>` | Displays the current user's numeric ID (disabled/info only) |
| ❌ Quit | Closes the application |

The tray **tooltip** also reflects live network status:
- `Network: UP` when online
- `Network: DOWN` when offline

---

## Project Structure

```/dev/null/text#L1-7
go-cli-monitor/
├── main.go                  # Application entry point and all core logic
├── green_circle_icon_32.png # Embedded tray icon — network UP
├── red_circle_icon_32.png   # Embedded tray icon — network DOWN
├── go.mod                   # Go module definition
└── go.sum                   # Dependency checksums
```

---

## Dependencies

| Package | Purpose |
|---------|---------|
| [`github.com/getlantern/systray`](https://github.com/getlantern/systray) v1.2.2 | Cross-platform system tray support |

All other entries in `go.sum` are transitive dependencies of `systray`.

---

## How It Works

1. **Startup** — If a path argument is provided, its existence is checked and reported immediately before the tray initializes.
2. **Tray setup** (`onReady`) — Icons are loaded from embedded bytes, hostname/UID are resolved, and static menu items are created.
3. **Monitor loop** — A background goroutine clears the terminal screen, prints current system info, performs an HTTP GET to `https://google.com` with a 5-second timeout, and updates the tray icon and tooltip accordingly. This loop repeats every 5 seconds.
4. **Quit handler** — A second goroutine waits for a click on the Quit menu item and calls `systray.Quit()`.
5. **Shutdown** (`onExit`) — Prints a confirmation message on clean exit.

---

## License

This project does not currently include a license file. Please contact the repository owner for usage terms.

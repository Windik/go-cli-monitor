# go-cli-monitor

A lightweight, cross-platform system monitor written in Go that lives in your system tray and provides real-time network status and basic system information directly in the terminal.

---

## Features

- 🟢🟠🔴 **Three-State System Tray Icon** — Green when all targets are reachable, orange when some are down, red when all are unreachable.
- 🌐 **Multi-Target Network Monitoring** — Concurrently checks all configured URLs in parallel; results are aggregated via Go channels.
- 📋 **Per-Target Tray Submenu** — A live "🌐 Network" submenu shows each target's status (🟢 UP / 🔴 DOWN) updated on every check cycle.
- ⚙️ **TOML Configuration** — All behaviour (targets, check interval, tray title) is driven by `config.toml`, which is **auto-created with defaults** on first run.
- 📝 **Structured File Logging** — Every significant event (errors, startup, shutdown) is appended to `monitor.log` with a timestamp and severity level (INFO / WARNING / ERROR / CRITICAL).
- 💻 **System Info** — Displays machine hostname and current user ID in both the terminal and the tray tooltip/menu.
- 📁 **Path Checker** — Accepts an optional CLI argument to verify whether a path exists, with distinct `[NOT FOUND]` vs `[ACCESS ERROR]` statuses.
- 🎨 **Colored Terminal Output** — ANSI color codes provide instant readability: green for UP, red for errors, etc.
- 🖥️ **Cross-Platform** — Supports macOS, Linux, and Windows (screen clearing uses `clear` or `cls` accordingly).
- 📦 **Self-Contained Binary** — All three tray icons are embedded at compile time using Go's `//go:embed` directive; no external assets needed at runtime.

---

## Requirements

- **Go** 1.21+ (module declares `go 1.26.1`)
- A desktop environment with system tray support (required by `getlantern/systray`)

---

## Installation

```/dev/null/sh#L1-3
git clone https://github.com/your-username/go-cli-monitor.git
cd go-cli-monitor
```

### Build

```/dev/null/sh#L1-1
go build -o go-cli-monitor .
```

### Run

```/dev/null/sh#L1-5
# Run the monitor (tray + terminal output)
./go-cli-monitor

# Run with a path argument to check its existence
./go-cli-monitor /path/to/check
```

> **First run:** if `config.toml` is absent, it will be created automatically with sensible defaults before the monitor starts.

---

## Configuration

All settings live in `config.toml` in the working directory.

```/dev/null/toml#L1-12
# Configuration file for go-cli-monitor

# Default title for the system tray
default_title = ""

# Check interval in seconds
check_interval = 5

# Monitoring targets (any number of HTTP/HTTPS URLs)
targets = [
    "https://google.com",
    "https://github.com",
    "https://sample-example-not-real.ru"
]
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `default_title` | string | `"CLI Monitor"` | Title shown in the system tray bar |
| `check_interval` | int | `5` | Seconds between each check cycle |
| `targets` | list of strings | `["https://google.com"]` | URLs to monitor; all are checked in parallel |

If `config.toml` is missing at startup, the application automatically writes a new file with the default values above and continues running.

---

## Usage

### Basic — System Monitor

Running the binary starts the monitor in both the terminal and the system tray:

```/dev/null/text#L1-6
=== System Monitor (Last update: 2024-06-01T12:00:05Z) ===

Hostname:       my-macbook.local
User ID:        501
Connection for url - https://google.com is      [UP] (Status: 200 OK)
Connection for url - https://github.com is      [UP] (Status: 200 OK)
Connection for url - https://sample-example-not-real.ru is      [DOWN] (Error: ...)
```

The terminal clears and refreshes every `check_interval` seconds. All target URLs are checked **concurrently** — the cycle duration is bounded by the slowest single response plus the 5-second HTTP timeout.

### System Tray Icon States

| Icon | Meaning |
|------|---------|
| 🟢 Green | All configured targets are reachable |
| 🟠 Orange | At least one target is down (partial outage) |
| 🔴 Red | All targets are unreachable |

### Path Argument

Pass a filesystem path as the first argument to check its status before the monitor starts:

```/dev/null/sh#L1-8
./go-cli-monitor /etc/hosts
# Path '/etc/hosts':    [EXISTS]

./go-cli-monitor /nonexistent/file
# Path '/nonexistent/file':    [NOT FOUND]

./go-cli-monitor /root/secret
# Path '/root/secret':  [ACCESS ERROR]
```

| Output | Meaning |
|--------|---------|
| `[EXISTS]` | Path was found and is accessible |
| `[NOT FOUND]` | Path does not exist |
| `[ACCESS ERROR]` | Path may exist but the current user lacks permission to stat it |

### Quitting

Click **❌ Quit** in the system tray menu to gracefully shut down the application.

---

## System Tray Menu

```/dev/null/text#L1-9
💻 Host: my-macbook.local        (info, disabled)
👤 User ID: 501                  (info, disabled)
─────────────────────────────
🌐 Network
    ├── 🟢 https://google.com
    ├── 🟢 https://github.com
    └── 🔴 https://sample-example-not-real.ru
─────────────────────────────
❌ Quit
```

The tray **tooltip** updates dynamically each cycle:
- `All 3 targets are UP | Host: my-macbook.local`
- `Warning: 2/3 targets UP | Host: my-macbook.local`
- `Critical: All targets are DOWN! | Host: my-macbook.local`

---

## Logging

All events are appended to `monitor.log` in the working directory. The log is never rotated or truncated automatically — manage it manually or via an external tool like `logrotate`.

**Format:**
```/dev/null/text#L1-4
2024-06-01 12:00:00 [INFO] Config loaded
2024-06-01 12:00:00 [INFO] No arguments provided.
2024-06-01 12:00:05 [ERROR] Connection for url - https://sample-example-not-real.ru [DOWN] - [ERROR] ...
2024-06-01 12:00:10 [INFO] Applications succesfuly closed
```

**Log levels:**

| Level | Usage |
|-------|-------|
| `INFO` | Normal lifecycle events (startup, shutdown, config loaded) |
| `WARNING` | Non-critical issues |
| `ERROR` | Failed network checks, path errors, icon loading failures |
| `CRITICAL` | Reserved for severe failures |

---

## Project Structure

```/dev/null/text#L1-12
go-cli-monitor/
├── main.go                        # Entry point — tray lifecycle, monitor loop, path checker
├── internal/
│   ├── config/
│   │   └── config.go              # TOML config loading via Viper (auto-creates defaults)
│   └── logger/
│       └── logger.go              # File-based structured logger
├── config.toml                    # Runtime configuration (auto-generated if absent)
├── monitor.log                    # Persistent log output (created at runtime)
├── green_circle_icon_32.png       # Embedded tray icon — all targets UP
├── orange_circle_icon_32.png      # Embedded tray icon — partial outage
├── red_circle_icon_32.png         # Embedded tray icon — all targets DOWN
├── go.mod                         # Go module definition
└── go.sum                         # Dependency checksums
```

---

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| [`github.com/getlantern/systray`](https://github.com/getlantern/systray) | v1.2.2 | Cross-platform system tray integration |
| [`github.com/spf13/viper`](https://github.com/spf13/viper) | v1.21.0 | TOML config loading, env binding, and default management |

All remaining entries in `go.sum` are transitive dependencies of the above two.

---

## How It Works

1. **Startup** — `config.toml` is loaded via Viper. If absent, a default file is written and used. If a CLI path argument is provided, its existence is checked immediately.
2. **Tray setup** (`onReady`) — The three PNG icons are read from embedded bytes. Hostname and UID are resolved. Static info items and a `🌐 Network` submenu are created — one sub-item per configured target.
3. **Monitor loop** — A background goroutine runs every `check_interval` seconds:
   - Clears the terminal and prints a header with the current timestamp.
   - Launches one goroutine per target; each makes an HTTP GET with a 5-second timeout and sends a `CheckResult` to a buffered channel.
   - The main loop reads all results, updates each submenu item's label (🟢 / 🔴), tallies up-count, and sets the tray icon and tooltip accordingly.
4. **Quit handler** — A second goroutine waits for a click on ❌ Quit and calls `systray.Quit()`.
5. **Shutdown** (`onExit`) — Prints a confirmation to stdout and writes a final `[INFO]` entry to `monitor.log`.

---

## License

This project does not currently include a license file. Please contact the repository owner for usage terms.

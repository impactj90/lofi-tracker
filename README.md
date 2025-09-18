# 🕰️ Lofi Tracker -- WIP --

Lofi Tracker is a minimal, Git-aware CLI time tracker built for developers who want to **stay in flow** and automatically track time on Jira tickets without breaking their rhythm.

Designed to be simple, branch-driven, and background-aware, it allows you to track time on the ticket you're working on — with **zero distractions**.

---

## ✨ Features

- ✅ Start/stop work tracking per Git branch (1 branch = 1 ticket)
- ✅ Auto-pauses when you're AFK (e.g., away for 10+ minutes)
- ✅ Auto-resumes when you're back
- ✅ Session summary and status
- ✅ SQLite-based persistent storage
- ✅ CLI and daemon architecture
- ✅ OS-native desktop notifications (Linux, macOS, Windows)
- ✅ Manual pause/resume support
- ✅ Daemon auto-switches sessions to match the current Git branch (detached HEAD ➜ `detached-<sha>`)
- ✅ Live status and summaries include in-progress session time
- ✅ CSV export for arbitrary date ranges
- 🧠 Tracks AFK vs. manual pauses separately
- 🧪 Reliable state transitions (AFK-aware state machine)

---

## 🛠 Installation

### 1. Install Go (if not yet installed)

```bash
https://go.dev/dl/
```

Ensure `go` is on your `$PATH`.

---

### 2. Clone and build the project

```bash
git clone https://github.com/<yourname>/lofi-tracker.git
cd lofi-tracker
make build
```

This will produce:

- `bin/lofi-tracker` – your main CLI
- `bin/lofi-daemon` – your AFK monitor

---

### 3. (Linux only) Install `notify-send` and `xprintidle`

These are needed for notifications and idle detection:

```bash
sudo apt install libnotify-bin xprintidle
```

> ⚠️ Not currently supported on Wayland (requires X11)

---

### 4. Install binaries (optional)

To make the CLI globally available:

```bash
sudo install bin/lofi-tracker /usr/local/bin/
sudo install bin/lofi-daemon /usr/local/bin/
```

Now you can use `lofi-tracker` and `lofi-daemon` from anywhere.

---

## 🚀 Usage

### ✅ Start your working day

```bash
cd your-project/
git checkout feature/ABC-123
lofi-tracker start
```

Tracks time on branch `feature/ABC-123`.

---

### ⏸ Pause or resume manually

```bash
lofi-tracker pause
lofi-tracker resume
```

---

### 🧘 Complete your working session

```bash
lofi-tracker complete
```

---

### 📊 Check your current status

```bash
lofi-tracker status
```

Shows the current branch, elapsed session time, and today's totals (including the live session).

---

### 🤝 Jump into meeting mode

```bash
lofi-tracker mm -b P2PMER-4
```

Force-completes the current session (if any) and starts tracking time on the specified meeting branch (defaults to `meeting`). Perfect for quick context switches when work isn't tied to a dev branch.

---

### 🧠 Background AFK detection (with OS notifications)

Daemonize activity tracker:

```bash
lofi-daemon
```

It:
- Checks idle time every 14 minutes
- Pauses your session if idle ≥ 14 minutes
- Resumes it when you return
- Sends OS notifications when paused/resumed
- Keeps sessions aligned with your current Git branch (completes the old one and starts tracking the new branch automatically)
- Falls back to `detached-<sha>` labels when you're working in a detached HEAD state

> ✅ Works silently in background, notifies you visually

---

### 📦 Export tracked time

```bash
lofi-tracker export --from 2024-10-01 --to 2024-10-31 --output work.csv
```

Writes `date,ticket,minutes` rows to CSV (defaults to the current month when flags are omitted). Leave `--output` empty to stream the CSV to stdout.

---

## 🧪 Testing

```bash
make test
```

Runs unit and integration tests.

---

## 📁 Data Storage

All tracking data is stored in:

```bash
~/.lofi-tracker/lofi-tracker.db
```

SQLite-powered. Portable. Inspectable.

---

## 🔔 Notifications Support

| OS       | Requires              |
|----------|-----------------------|
| macOS    | Nothing (uses AppleScript) |
| Linux    | `notify-send`, install with `sudo apt install libnotify-bin` |
| Windows  | Built-in (via native APIs) |

Uses [`github.com/gen2brain/beeep`](https://github.com/gen2brain/beeep) under the hood.

---

## 💡 Roadmap Ideas

- [ ] Daily/weekly/monthly summaries
- [ ] Reminder to start your working day
- [ ] Manual correction (add time retroactively)
- [ ] Jira time sync
- [ ] `daemon start`/`stop` command via PID file
- [ ] `--no-notify` flag or config toggle

---

## ❤️ Contributing

Pull requests are welcome. If you’ve got an idea, issue, or enhancement, open it!

---

Lofi Tracker is built to remove friction from time tracking — not to add it.

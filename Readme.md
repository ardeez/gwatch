# gwatch

A general-purpose file watcher and hot-reload tool for Go projects. Detects file changes via polling and automatically rebuilds and restarts your binary — without you having to re-run anything manually.

---

## Functional Requirements

### CLI Usage

```bash
gwatch -entry ./cmd/server -dir . -ext .go -exclude vendor,tmp -interval 500
```

### Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `-entry` | string | *(required)* | Entry point to build, e.g. `./cmd/server` |
| `-dir` | string | `.` | Root directory to watch |
| `-ext` | string | `.go` | File extension to watch |
| `-exclude` | string | `vendor,tmp` | Comma-separated directories to exclude |
| `-interval` | int | `500` | Poll interval in milliseconds |

### Behavior

| # | Requirement |
|---|---|
| FR-1 | Tool is general-purpose — usable in any Go project, not tied to a specific structure |
| FR-2 | User can specify entry point via `-entry` flag; no hardcoded path assumption |
| FR-3 | Watcher polls the filesystem every `-interval` ms using `os.Stat()` to read `modTime` |
| FR-4 | Changes are detected by comparing current `modTime` against an in-memory snapshot (`map[string]time.Time`) |
| FR-5 | Debounce of 300ms is applied after a change is detected — timer resets on each subsequent save before firing |
| FR-6 | After debounce timer expires, an empty signal (`struct{}{}`) is sent to a channel to trigger rebuild |
| FR-7 | Runner receives signal from channel and runs `go build -o ./tmp/app <entry>` |
| FR-8 | Build output binary is always written to `./tmp/` |
| FR-9 | If build succeeds — kill old process, start new binary |
| FR-10 | If build fails — log error to terminal, keep old process running |
| FR-11 | On startup, runner performs an initial build and run before any file change is detected |
| FR-12 | Directories listed in `-exclude` are skipped during polling |

---

## Quality Attributes

### Reliability
- Watcher goroutine uses `recover()` to catch panics and restart the internal poll loop instead of crashing the entire tool
- If build fails, the previously running process is preserved — development is never interrupted by a typo

### Performance
- Designed for small-to-medium Go projects (under ~1000 files)
- Polling via `os.Stat()` is lightweight; no OS-level event subscription required
- Debounce prevents redundant rebuilds caused by rapid consecutive saves

### Portability
- Compatible with Linux, macOS, and Windows
- Uses only Go standard library primitives (`os`, `os/exec`, `time`, `flag`) — no OS-specific syscalls

### Observability
- Structured log levels: `INFO`, `WARN`, `ERROR`, `DEBUG`
- Each log line includes a timestamp
- Colored terminal output per log level for fast visual scanning

### Usability
- Single binary, zero config file required
- All options configurable via CLI flags with sensible defaults
- On build failure, full compiler error is printed to terminal without stopping the watcher

---

## Sequence Diagram

```
main.go        watcher        debounce       channel        runner         filesystem
   |               |               |              |              |               |
   |--start(config)--------------->|              |              |               |
   |               |               |              |              |               |
   |--start runner goroutine---------------------------------->  |               |
   |               |               |              |              |               |
   |               |               |              |   initial build + run------->|
   |               |               |              |              |<---binary ready
   |               |               |              |              |               |
   |    [POLL LOOP every 500ms]    |              |              |               |
   |               |--os.Stat() all .go files--------------------------->        |
   |               |<--modTime per file----------------------------------------|
   |               |               |              |              |               |
   |               |--compare snapshot             |              |               |
   |               |               |              |              |               |
   |            [changed?]         |              |              |               |
   |               |               |              |              |               |
   |    no → wait 500ms → repeat   |              |              |               |
   |               |               |              |              |               |
   |    yes        |               |              |              |               |
   |               |--reset timer 300ms-->|        |              |               |
   |               |               |              |              |               |
   |            [more saves?]      |              |              |               |
   |               |    yes → reset timer          |              |               |
   |               |               |              |              |               |
   |               |   timer expired              |              |               |
   |               |               |--chan struct{}{}-->|         |               |
   |               |               |              |              |               |
   |    [REBUILD]  |               |              |              |               |
   |               |               |              |   go build -o ./tmp/app----->|
   |               |               |              |              |<---result      |
   |               |               |              |              |               |
   |            [build ok?]        |              |              |               |
   |               |               |              |              |               |
   |    ok  → kill old proc → run new binary      |              |               |
   |    fail → log error, keep old process running |              |               |
```

---

## Component Overview

```
gwatch/
├── main.go        — flag parsing, wires all components together
├── watcher.go     — poll loop, snapshot comparison, triggers debounce
├── debounce.go    — 300ms debounce timer, sends signal to channel
└── runner.go      — receives signal, builds binary, manages process lifecycle
```

---

## Internal Data Flow

```
watcher.go
  └── polls filesystem every 500ms
  └── detects modTime change
  └── calls debounce

debounce.go
  └── resets 300ms timer on each call
  └── on expiry → sends struct{}{} to chan

runner.go
  └── blocks on <-chan
  └── runs go build
  └── on success: kills old *exec.Cmd, starts new process
  └── on failure: logs error, keeps old process alive
```
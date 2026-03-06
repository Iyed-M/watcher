# watcher

A distributed, extensible activity tracking system designed to capture precise user context across desktop (Linux, Windows) and mobile (Android) environments.

Unlike basic screen-time trackers, `watcher` understands *what* you are doing, not just what application is open (e.g., `12:00 -> 13:00 coding in neovim on project 'book_app' focusing on feature 'feature_x' on git_branch 'feat/feature_x/refactor'`, `14:00 -> 14:30 watching youtube video 'the_title'`).

## Architecture

Watcher is split into three decoupled layers to ensure plugins remain lightweight, daemons handle network instability gracefully, and the central server reliably resolves cross-device conflicts.

* **Central Server:** The global source of truth. Receives batched data, applies user-defined priority rules to resolve overlapping timelines (e.g., Mobile Active > Desktop Idle), and serves the frontend.
* **Device Daemons:** Background processes running on each OS. They aggregate local watcher data, manage offline SQLite caching, and handle batched network synchronization.
* **Watchers:** Independent, highly-scoped plugins residing inside host applications (Neovim, VSCode, Firefox) or window managers (Hyprland). They are strictly event-driven (push-only) and consume zero CPU when idle.

## Monorepo Structure

```text
.
├── server/                  Go backend & Postgres rules engine
├── daemons/
│   ├── desktop/             Go daemon for Linux/Windows (UDS & WebSocket listeners)
│   └── android/             Kotlin Foreground Service (UsageStatsManager & WorkManager)
├── watchers/
│   ├── hyprland/            Bash/Go script hooking into Hyprland socket2
│   ├── neovim/              Lua plugin using vim.uv
│   ├── vscode/              TypeScript extension
│   └── browser/             JS extension (Firefox/Chrome) via WebSockets/Native Messaging
├── shared/                  Shared protobuf definitions & JSON schemas
└── docs/                    Architecture and API documentation

```

## Tech Stack

* **Server:** Go, PostgreSQL, gRPC/REST
* **Desktop Daemon:** Go, SQLite, Unix Domain Sockets (UDS), WebSockets
* **Android Daemon:** Kotlin, Room (SQLite), WorkManager, UsageEvents API
* **Watchers:** Lua, TypeScript, JavaScript, Bash, Go

## Writing a Custom Watcher

Watchers are language-agnostic. Any script that can connect to the daemon's local socket (UDS on Linux, Named Pipes on Windows) and stream JSON Lines (JSONL) can act as a watcher.

**Expected Payload Format:**

```json
{
  "watcher_id": "custom-script",
  "process_id": 1234,
  "timestamp": "2026-03-06T01:54:50Z",
  "event_type": "state_change",
  "app_name": "my-app",
  "context": { "project": "foo", "task": "bar" }
}

```

## Getting Started

WIP

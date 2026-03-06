# Watcher Project: Architecture & Contribution Guide

Welcome to the Watcher monorepo. This document outlines our core engineering philosophy, system architecture, and how the various components communicate. Before contributing, please read this entirely to understand our design constraints and goals.

## 1. Core Engineering Philosophy

We are building a highly performant, privacy-first, cross-platform time and context tracker. Every line of code must adhere to these three principles:

1. **Event-Driven Only:** Polling wastes CPU and battery. Every component must sleep until the OS or a host application wakes it up with a hardware interrupt or event hook.
2. **Dumb Watchers, Smart Daemons, Omniscient Server:** * Watchers contain zero logic; they blindly push formatted strings.
    * Local Daemons resolve OS-level process trees and handle offline caching.
    * The Server resolves cross-device conflicts to build the final timeline.
3. **Dependency Injection (100% Testability):** Core business logic never directly touches sockets, databases, or the network. We pass interfaces so every module can be unit-tested using mock memory buffers.

---

## 2. System Architecture Overview

The system is split into three distinct layers, isolating context extraction from data persistence and global resolution.

### A. The Watchers (Data Extractors)

Watchers are lightweight programs or plugins living inside host applications (like Neovim or VSCode) or the window manager.

* **Responsibility:** Hook into native application events, extract the active context (e.g., current file, git branch, window title), format it as JSON Lines, and push it to the local IPC socket.
* **Implementations:**
  * **Root (Hyprland/Linux):** Go script reading `.socket2.sock`. Acts as the global source of truth for window focus. **WIP**
  * **App (Neovim):** Lua plugin utilizing `vim.uv` named pipes triggered by `autocmd`. **NOT YET IMPLEMENTED**
  * **App (VSCode):** TypeScript extension utilizing the Node `net` module on `onDidChangeActiveTextEditor`. **NOT YET IMPLEMENTED**
  * **App (Browser):** JS background worker using WebSockets to `localhost`. **NOT YET IMPLEMENTED**

### B. The Local Daemons (Data Aggregators) **NOT YET IMPLEMENTED**

The daemon is an offline-first background process running on the user's device.

* **Responsibility:** Orchestrate local watchers, read OS hardware interrupts for idle states, resolve local process tree conflicts (e.g., is Neovim in the foreground or background?), and securely cache data offline.
* **Tech Stack:** Go (Desktop) / Kotlin (Android).
* **The Desktop Pipeline (Go):**

    1. **IPC Layer:** Uses `//go:build` tags to abstract Unix Domain Sockets (Linux) and Named Pipes (Windows). Routes incoming JSON to a central Go channel.
    2. **Idle Monitor:** Detects user AFK states via `/dev/input` (Linux) or `GetLastInputInfo` (Windows).
    3. **Aggregator:** Reads the event channel, debounces rapid context switching, validates child-parent process relationships, and finalizes discrete timeline blocks.
    4. **Cache:** Stores blocks durably using embedded SQLite (`modernc.org/sqlite`).
    5. **Sync Engine:** Periodically wakes up, queries SQLite for unsynced rows, and POSTs batches to the server.

### C. The Central Server (Global Resolver) **NOT_YET_IMPLENETE

The central source of truth for a user's cross-device activity.

* **Responsibility:** Ingest batched data from multiple devices, apply conflict resolution rules (e.g., mobile activity overrides desktop idle time), and serve the compiled timeline to the frontend.
* **Tech Stack:** Go, PostgreSQL.

---

## 3. Communication Protocols & Schemas

To ensure compatibility across Go, TypeScript, Lua, and Kotlin, all schemas are defined in `/shared/schemas/`.

### Local IPC (Watcher -> Daemon)

* **Protocol:** JSONL (JSON Lines) streamed over UDS, Named Pipes, or WebSockets.
* **Payload Format:**

  ```json
  {
    "watcher_id": "neovim",
    "process_id": 4512,
    "timestamp": "2026-03-06T20:21:11Z",
    "event_type": "state_change",
    "app_name": "neovim",
    "context": { "project": "watcher_app", "file": "main.go" }
  }
  ```

  **context** can contain any arbitrary data that gives more context to the activity
  (e.g. youtube_video_title, git_branch, ... )

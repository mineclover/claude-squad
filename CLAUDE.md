# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Go Application (Main)
- `go build` - Build the main application
- `go test ./...` - Run all Go tests
- `go run main.go` - Run the application locally
- `go run main.go debug` - Show debug info and config paths

### Web Frontend (web/ directory)
- `cd web && npm run dev` - Start development server with Turbopack
- `cd web && npm run build` - Build for production
- `cd web && npm run lint` - Run ESLint

## Architecture Overview

Claude Squad is a terminal application that manages multiple AI assistants (Claude Code, Aider, Codex, etc.) in isolated workspaces.

### Key Components

**Core Application Structure:**
- `main.go` - Entry point with CLI commands using Cobra
- `app/` - Main application logic with Bubble Tea TUI framework
- `session/` - Instance management and storage
- `ui/` - Terminal UI components (list, menu, preview, diff views)

**Session Management:**
- Each AI assistant instance runs in its own tmux session for isolation
- Git worktrees are used to create separate branches per instance
- Instances can be paused/resumed and changes can be committed/pushed

**Storage & Configuration:**
- `config/` - Configuration and persistent state management
- `session/storage.go` - Instance persistence across app restarts
- State stored in user config directory (use `cs debug` to see path)

**Git Integration:**
- `session/git/` - Git worktree and branch management
- Each instance works on its own branch to avoid conflicts
- Built-in support for committing and pushing changes

**tmux Integration:**
- `session/tmux/` - tmux session management for process isolation
- Handles PTY creation and terminal attachment

### Key Features
- Background task execution with auto-accept mode (`-y` flag)
- Real-time diff viewing and change preview
- Instance isolation using git worktrees
- Support for multiple AI assistants via `-p` flag

## Prerequisites
- tmux (required for session isolation)
- gh CLI (required for GitHub integration)
- Git repository (must run from within git repo)

## Configuration
- Default program: `claude` (Claude Code)
- Config location: `~/.config/claude-squad/` or use `cs debug` to find
- Auto-yes mode available for background execution

## Testing
- Go tests: Use `go test ./...` for all packages
- Web tests: Run from `web/` directory with `npm run lint`
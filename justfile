# List all recipes
default:
    @just --list

# Repo name used as tmux session prefix (auto-detected from directory name)
repo := `basename $(git rev-parse --show-toplevel 2>/dev/null || basename $PWD)`
# Random 4-digit suffix to allow multiple sessions of the same type
rand := `printf '%04d' $((RANDOM % 10000))`

# Start a Gemini session in YOLO approval mode
g_session:
    @gemini --approval-mode yolo

# Start a Claude session (default model) in tmux
c_session:
    @claude --dangerously-skip-permissions --model 'opus' "/caveman"

# Start Claude with Sonnet model in tmux
sonnet:
    @tmux new-session -d -s {{repo}}-sonnet-{{rand}} 'claude --dangerously-skip-permissions --model "sonnet"' && tmux attach -t {{repo}}-sonnet-{{rand}}

# Start Claude with Opus model in tmux
opus:
    claude --dangerously-skip-permissions

# Start Claude with Haiku model in tmux
haiku:
    @tmux new-session -d -s {{repo}}-haiku-{{rand}} 'claude --dangerously-skip-permissions --model "haiku"' && tmux attach -t {{repo}}-haiku-{{rand}}

dev: build-helper
    PATH="$HOME/go/bin:$PATH" wails dev

r_mix:
    @repomix --parsable-style --compress --remove-empty-lines --skill-generate use-repo-code

# Build and launch Mashed
run: build
    open build/bin/mashed.app

# Build and sign the PTY helper binary (entitlements required for PTY on macOS Sequoia)
build-helper:
    go build -o build/bin/mashed-pty-helper ./cmd/pty-helper
    codesign --force --options runtime --sign "Apple Development: linus McManamey (5X8A9U965U)" --entitlements build/darwin/entitlements.plist build/bin/mashed-pty-helper

# Full build: helper + wails + bundle + sign
build: build-helper
    cd frontend && npm install && cd ..
    PATH="$HOME/go/bin:$PATH" wails build
    cp -r fonts build/bin/mashed.app/Contents/Resources/fonts
    cp build/bin/mashed-pty-helper "build/bin/mashed.app/Contents/MacOS/mashed-pty-helper"
    codesign --force --options runtime --sign "Apple Development: linus McManamey (5X8A9U965U)" --entitlements build/darwin/entitlements.plist "build/bin/mashed.app/Contents/MacOS/mashed-pty-helper"
    codesign --force --options runtime --sign "Apple Development: linus McManamey (5X8A9U965U)" --entitlements build/darwin/entitlements.plist "build/bin/mashed.app"

# Run Go tests
test:
    go test ./internal/... -count=1


# List all tmux sessions related to this repo, grouped by parent/child
sessions:
    #!/usr/bin/env bash
    repo=$(basename "$(git rev-parse --show-toplevel 2>/dev/null || echo "$PWD")")
    echo "tmux sessions for: $repo"
    echo "════════════════════════════════════════════════"
    # Parent sessions: match repo name prefix
    parents=$(tmux list-sessions -F '#{session_name}|#{session_created}|#{session_attached}|#{session_windows}' 2>/dev/null | grep "^${repo}" || true)
    # Child sessions: common agent prefixes (surfseer, bmad, clawteam, term)
    children=$(tmux list-sessions -F '#{session_name}|#{session_created}|#{session_attached}|#{session_windows}' 2>/dev/null | grep -E "^(surfseer|bmad|clawteam|term)-" || true)
    if [ -z "$parents" ] && [ -z "$children" ]; then
        echo "  No sessions found."
        exit 0
    fi
    if [ -n "$parents" ]; then
        echo ""
        echo "  Parent sessions (${repo}-*):"
        echo "$parents" | while IFS='|' read -r name created attached windows; do
            ts=$(date -r "$created" '+%H:%M' 2>/dev/null || echo "$created")
            status=""
            [ "$attached" = "1" ] && status=" ◀ attached"
            printf "    %-40s %s  (%s win)%s\n" "$name" "$ts" "$windows" "$status"
        done
    fi
    if [ -n "$children" ]; then
        echo ""
        echo "  Agent sessions (surfseer-*, bmad-*, clawteam-*, term-*):"
        echo "$children" | while IFS='|' read -r name created attached windows; do
            ts=$(date -r "$created" '+%H:%M' 2>/dev/null || echo "$created")
            status=""
            [ "$attached" = "1" ] && status=" ◀ attached"
            printf "    %-40s %s  (%s win)%s\n" "$name" "$ts" "$windows" "$status"
        done
    fi
    echo ""
    p_count=0; [ -n "$parents" ] && p_count=$(echo "$parents" | wc -l | tr -d ' ')
    c_count=0; [ -n "$children" ] && c_count=$(echo "$children" | wc -l | tr -d ' ')
    total=$((p_count + c_count))
    echo "  Total: $total sessions"

# Tile all bmad-surfseer-* tmux sessions into a single viewer window
view-bmad:
    @bash scripts/tmux-bmad-viewer.sh

# Attach to all repo-related tmux sessions in a single tmux window (switch with prefix+s)
attach-all:
    #!/usr/bin/env bash
    repo=$(basename "$(git rev-parse --show-toplevel 2>/dev/null || echo "$PWD")")
    sessions=$(tmux list-sessions -F '#{session_name}' 2>/dev/null | grep -E "^(${repo}|surfseer|bmad|clawteam|term)-" || true)
    if [ -z "$sessions" ]; then
        echo "No repo sessions found."
        exit 0
    fi
    count=$(echo "$sessions" | wc -l | tr -d ' ')
    first=$(echo "$sessions" | head -1)
    echo "Found $count sessions. Attaching to: $first"
    echo "Use prefix+s to switch between sessions, or prefix+( / prefix+) to cycle."
    tmux attach -t "$first"

# Check for Claude Code updates and install if available
update:
    @claude --version && claude update

# Alias for update
upgrade: update

repomixer:
    # repomix --remote https://github.com/peteromallet/desloppify --compress -o ./docs/repomixer/desloppify/desloppify.xml --style xml
    # repomix --remote https://github.com/wailsapp/wails  --compress -o ./docs/repomixer/wails/wails.xml --style xml
    # repomix --remote https://github.com/xyflow/xyflow  --compress -o ./docs/repomixer/xyflow/xyflow.xml --style xml
    # repomix --remote https://github.com/bmad-code-org/BMAD-METHOD  --compress -o ./docs/repomixer/bmad-method/bmad-method.xml --style xml
    repomix --remote https://github.com/manaflow-ai/cmux  --compress -o ./docs/repomixer/cmux/cmux.xml --style xml   

# Run Go tests then frontend tests
test-all: test
    cd frontend && npx vitest run

# Run per-package Go coverage threshold enforcement
test-cover:
    bash scripts/check-coverage.sh

# Run Go static analysis
lint:
    go vet ./...

# Install lefthook git hooks
hooks-install:
    lefthook install --force

# Manually run pre-commit hooks without committing
pre-check:
    lefthook run pre-commit


# Claude tmux Swarm Discovery
# Usage: just <recipe>



# ── Session & Pane Enumeration ─────────────────────────────────────────────

# List all active tmux sessions
list_sessions:
    tmux list-sessions

# List all panes across all sessions with running command and PID
panes:
    tmux list-panes -a -F \
        '#{session_name}:#{window_index}.#{pane_index} | cmd=#{pane_current_command} | pid=#{pane_pid}'

# Full pane inventory — all fields useful for Claude Conductor
panes-full:
    tmux list-panes -a -F \
        '#{session_id}|#{session_name}|#{window_index}|#{window_name}|#{pane_index}|#{pane_id}|#{pane_current_command}|#{pane_pid}|#{pane_title}'

# ── Claude Detection ───────────────────────────────────────────────────────

# Tier 1: Find panes where pane_current_command is directly 'claude'
detect-direct:
    @tmux list-panes -a -F \
        '#{session_name}:#{window_index}.#{pane_index} | pid=#{pane_pid} | cmd=#{pane_current_command}' \
    | grep ' cmd=claude'

# Tier 2: Find claude processes via process tree (catches zsh-wrapped launches)
detect-proctree:
    @for pid in $(pgrep -x claude 2>/dev/null); do \
        echo "claude PID: $pid"; \
        tmux list-panes -a -F '#{pane_pid} #{pane_id} #{session_name}:#{window_index}.#{pane_index}' \
        | awk -v pid="$pid" '$1 == pid {print "  tmux pane: " $2 " (" $3 ")"}'; \
    done

# Tier 3: Find panes whose visible output contains Claude Code UI markers
detect-content:
    @tmux list-panes -a -F '#{pane_id} #{session_name}:#{window_index}.#{pane_index}' \
    | while read pane_id addr; do \
        if tmux capture-pane -t "$pane_id" -p 2>/dev/null | grep -q "Claude\|claude>"; then \
            echo "$pane_id  $addr"; \
        fi; \
    done

# Run all three detection tiers
detect-all:
    @echo "=== Tier 1: Direct command match ==="
    @just detect-direct || true
    @echo ""
    @echo "=== Tier 2: Process tree ==="
    @just detect-proctree || true
    @echo ""
    @echo "=== Tier 3: Content capture ==="
    @just detect-content || true

# One-liner: filter panes running claude or node (agent wrapper)
discover:
    @tmux list-panes -a -F \
        '#{session_name}|#{window_index}|#{pane_index}|#{pane_id}|#{pane_current_command}|#{pane_pid}' \
    | awk -F'|' '$5 ~ /claude|node/ {print $1 ":" $2 "." $3 " | pane=" $4 " | cmd=" $5 " | pid=" $6}'

# ── Pane Output Capture ────────────────────────────────────────────────────

# Capture last 50 lines from a pane — usage: just capture PANE_ID
# e.g. just capture %3
capture PANE_ID:
    tmux capture-pane -t {{PANE_ID}} -p -S -50

# Capture last 100 lines from a pane
capture-100 PANE_ID:
    tmux capture-pane -t {{PANE_ID}} -p -S -100

# Capture last 200 lines from a pane to a temp file
capture-file PANE_ID:
    tmux capture-pane -t {{PANE_ID}} -p -S -200 > /tmp/agent-{{PANE_ID}}-output.txt
    @echo "Saved to /tmp/agent-{{PANE_ID}}-output.txt"

# ── Agent Team Config ──────────────────────────────────────────────────────

# List all Claude agent team configs
teams:
    @ls ~/.claude/teams/ 2>/dev/null || echo "No teams directory found"

# Read a specific team config — usage: just team-config TEAM_NAME
team-config TEAM_NAME:
    cat ~/.claude/teams/{{TEAM_NAME}}/config.json

# List task files for a team — usage: just team-tasks TEAM_NAME
team-tasks TEAM_NAME:
    @ls ~/.claude/tasks/{{TEAM_NAME}}/ 2>/dev/null || echo "No tasks found for {{TEAM_NAME}}"

# ── Environment & Context ──────────────────────────────────────────────────

# Show Claude-related env vars in current/target session
env-check:
    @echo "TMUX: ${TMUX:-not set}"
    @echo "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS: ${CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS:-not set}"
    @echo "CLAUDE_CODE_SPAWN_BACKEND: ${CLAUDE_CODE_SPAWN_BACKEND:-not set}"

# Check if agent teams env var is set in a target session — usage: just env-session SESSION
env-session SESSION:
    tmux show-environment -t {{SESSION}} CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS 2>/dev/null \
        || echo "Not set in session '{{SESSION}}'"

# Show all pane PIDs mapped to their tmux address
pid-map:
    @tmux list-panes -a -F '#{pane_pid} -> #{session_name}:#{window_index}.#{pane_index} (#{pane_id})'

# ── Process-Level Discovery ────────────────────────────────────────────────

# All claude processes on the system
procs:
    @pgrep -a claude 2>/dev/null || ps aux | grep -i claude | grep -v grep

# Map all claude PIDs back to their tmux panes
pid-to-pane:
    @for pid in $(pgrep -x claude 2>/dev/null); do \
        match=$(tmux list-panes -a \
            -F '#{pane_pid} #{pane_id} #{session_name}:#{window_index}.#{pane_index}' \
            | awk -v p="$pid" '$1 == p'); \
        if [ -n "$match" ]; then \
            echo "PID $pid => $match"; \
        else \
            echo "PID $pid => (no matching tmux pane)"; \
        fi; \
    done

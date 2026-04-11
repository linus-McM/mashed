#!/usr/bin/env bash
# Launch all bmad-surfseer tmux sessions in a single terminal view.
# Creates a new tmux session "bmad-viewer" with one pane per existing
# bmad-surfseer-* session, each attached read-write via nested tmux.

set -euo pipefail

VIEWER="bmad-viewer"

# Discover all bmad-surfseer-* sessions (portable for bash 3.2 on macOS)
SESSIONS=()
while IFS= read -r line; do
  [[ -n "$line" ]] && SESSIONS+=("$line")
done < <(tmux list-sessions -F '#{session_name}' 2>/dev/null | grep '^bmad-surfseer-' || true)

if [[ ${#SESSIONS[@]} -eq 0 ]]; then
  echo "No bmad-surfseer-* tmux sessions found." >&2
  exit 1
fi

# If viewer already exists, just attach
if tmux has-session -t "$VIEWER" 2>/dev/null; then
  exec tmux attach -t "$VIEWER"
fi

# Create the viewer session with the first sub-session in pane 0
tmux new-session -d -s "$VIEWER" -n bmad \
  "TMUX= tmux attach -t ${SESSIONS[0]}"

# Add a pane for each remaining session, splitting evenly
for ((i = 1; i < ${#SESSIONS[@]}; i++)); do
  tmux split-window -t "$VIEWER":0 -h \
    "TMUX= tmux attach -t ${SESSIONS[$i]}"
  tmux select-layout -t "$VIEWER":0 even-horizontal
done

tmux select-layout -t "$VIEWER":0 tiled
tmux select-pane -t "$VIEWER":0.0

exec tmux attach -t "$VIEWER"

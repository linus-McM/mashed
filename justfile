
# Start a Gemini session in YOLO approval mode
g_session:
    gemini --approval-mode yolo

# Start a Claude session (default model)
c_session:
    @claude --dangerously-skip-permissions

# Start Claude with Sonnet model
sonnet:
    @claude --dangerously-skip-permissions --model "sonnet"

# Start Claude with Opus model
opus:
    @claude --dangerously-skip-permissions

# Start Claude with Haiku model
haiku:
    @claude --dangerously-skip-permissions --model "haiku"

# Launch Mashed in dev mode (hot reload)
dev:
    PATH="$HOME/go/bin:$PATH" wails dev

# Build and launch Mashed
run: build
    open build/bin/conductor.app

# Build Mashed production binary
build:
    cd frontend && npm install && cd ..
    PATH="$HOME/go/bin:$PATH" wails build

# Run Go tests
test:
    go test ./internal/... -count=1


repomixer:
    repomix --remote https://github.com/peteromallet/desloppify --compress -o ./docs/repomixer/desloppify/desloppify.xml --style xml
    repomix --remote https://github.com/wailsapp/wails  --compress -o ./docs/repomixer/wails/wails.xml --style xml 
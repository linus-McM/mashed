
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

# Launch Claude Conductor dashboard
dev:
    bun run dev


repomixer:
    repomix --remote https://github.com/peteromallet/desloppify --compress -o ./docs/repomixer/desloppify/desloppify.xml --style xml
    repomix --remote https://github.com/wailsapp/wails  --compress -o ./docs/repomixer/wails/wails.xml --style xml 
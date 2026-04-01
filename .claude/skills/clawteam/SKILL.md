---
name: ClawTeam Multi-Agent Coordination
description: >
  This skill should be used when the user asks to "create a team", "spawn agents",
  "assign tasks", "coordinate multiple agents", "check team status", "view kanban board",
  "send messages between agents", "manage team tasks", "monitor team progress",
  or mentions "clawteam", "multi-agent coordination", "team collaboration",
  "agent inbox", "task board", "spawn worker". This skill should also be triggered
  when the current task is complex enough to benefit from splitting into subtasks
  and delegating to multiple agents — for example when the user asks to "build a
  full-stack app", "refactor the entire codebase", "implement multiple features
  in parallel", or when the agent determines that the work scope exceeds what a
  single agent can efficiently handle alone. Provides comprehensive guidance for
  using the ClawTeam CLI to orchestrate multi-agent teams with task management,
  messaging, and monitoring. Includes integration with 35 golang development skills
  for coordinating Go-focused teams — covers architecture, error handling, concurrency,
  testing, observability, CI/CD, and the samber ecosystem.
---

# ClawTeam Multi-Agent Coordination

ClawTeam is a framework-agnostic CLI tool for coordinating multiple AI agents as a team.
It provides file-based team management, inter-agent messaging, shared task tracking with
dependency resolution, plan approval workflows, and terminal-based monitoring dashboards.

All operations are performed via the `clawteam` CLI. Data is stored in `~/.clawteam/` by default.

## Installation

```bash
pip install clawteam
```

Requires Python 3.10+. For P2P transport support: `pip install clawteam[p2p]`.

## Prerequisites

- `tmux` installed (used by default spawn backend)
- A CLI coding agent (e.g. `claude`, `codex`, or any command-line agent)
- A git repository (for worktree isolation)

## Core Concepts

**Teams** — A named group of agents with one leader and zero or more workers. Created via
`clawteam team spawn-team`. The leader approves joins, reviews plans, and coordinates shutdown.

**Inbox** — File-based message queue per agent. `inbox send` for point-to-point, `inbox broadcast`
for all members. `inbox receive` consumes messages (destructive); `inbox peek` reads without consuming.

**Tasks** — Shared task board with statuses: `pending`, `in_progress`, `completed`, `blocked`.
Tasks support dependency chains (`--blocks`, `--blocked-by`). Completing a task auto-unblocks dependents.

**Board** — Terminal kanban dashboard. `board show` for single team, `board overview` for all teams,
`board live` for real-time auto-refresh, `board attach` for tiled tmux view of all agents.

**Identity** — Each agent has env vars (`CLAWTEAM_AGENT_ID`, `CLAWTEAM_AGENT_NAME`, `CLAWTEAM_AGENT_TYPE`,
`CLAWTEAM_TEAM_NAME`). Set automatically when spawned via `clawteam spawn`.

## Quick Start

### Set Up a Team with Tasks

```bash
# Set identity for the current session
export CLAWTEAM_AGENT_ID="leader-001"
export CLAWTEAM_AGENT_NAME="leader"
export CLAWTEAM_AGENT_TYPE="leader"

# Create team
clawteam team spawn-team my-team -d "Project team" -n leader

# Create tasks
clawteam task create my-team "Design system" -o leader
clawteam task create my-team "Implement feature" -o worker1
clawteam task create my-team "Write tests" -o worker2

# View board
clawteam board show my-team
```

### Spawn and Coordinate Agents

```bash
# Spawn workers — defaults: tmux backend, claude command, git worktree isolation, skip-permissions on
clawteam spawn --team my-team --agent-name worker1 --task "Implement the auth module"
clawteam spawn --team my-team --agent-name worker2 --task "Write unit tests"

# Or explicitly specify backend and command (positional args: [BACKEND] [COMMAND])
clawteam spawn tmux claude --team my-team --agent-name worker3 --task "Build API endpoints"
clawteam spawn subprocess claude --team my-team --agent-name worker4 --task "Run linting"

# Watch all agents working simultaneously (tiled tmux panes)
clawteam board attach my-team

# Send instructions
clawteam inbox send my-team worker1 "Start implementing the auth module"

# Monitor task board
clawteam board live my-team --interval 3
```

### Spawn Defaults

Spawning agents uses sensible defaults — no flags needed for the common case:

| Setting | Default | Override |
|---------|---------|----------|
| Backend | `tmux` | `clawteam spawn subprocess ...` |
| Command | `claude` | `clawteam spawn tmux my-cmd ...` |
| Workspace | `auto` (git worktree) | `--no-workspace` or config `workspace=never` |
| Permissions | skip (no approval needed) | `--no-skip-permissions` or config `skip_permissions=false` |

Agents spawned with defaults get:
- Their own **git worktree** (isolated branch, no conflicts with other agents)
- **Full tool permissions** (`--dangerously-skip-permissions`) so they can work autonomously
- A **tmux window** you can watch with `board attach`

### Task Lifecycle

```bash
# Create with dependencies
clawteam task create my-team "Deploy" --blocked-by <impl-task-id>,<test-task-id>

# Update status
clawteam task update my-team <task-id> --status in_progress
clawteam task update my-team <task-id> --status completed  # auto-unblocks dependents

# Filter tasks
clawteam task list my-team --status blocked
clawteam task list my-team --owner worker1
```

### Waiting for Sub-Agents

```bash
# Block until all tasks complete (no timeout)
clawteam task wait my-team

# With timeout and custom poll interval
clawteam task wait my-team --timeout 300 --poll-interval 10

# Monitor a specific agent's inbox instead of the leader
clawteam task wait my-team --agent coordinator

# JSON streaming output (NDJSON: progress + message events, then final result)
clawteam --json task wait my-team --timeout 600
```

### Watching Agents Work

```bash
# Tile all agent tmux windows into one view (best way to observe)
clawteam board attach my-team

# Or attach to the tmux session manually and switch windows with Ctrl-b + number
tmux attach -t clawteam-my-team
```

## Command Groups

| Group | Purpose | Key Commands |
|-------|---------|-------------|
| `team` | Team lifecycle | `spawn-team`, `discover`, `status`, `request-join`, `approve-join`, `cleanup` |
| `inbox` | Messaging | `send`, `broadcast`, `receive`, `peek`, `watch` |
| `task` | Task management | `create`, `get`, `update`, `list`, `wait` |
| `board` | Monitoring | `show`, `overview`, `live`, `attach`, `serve` |
| `plan` | Plan approval | `submit`, `approve`, `reject` |
| `lifecycle` | Agent lifecycle | `request-shutdown`, `approve-shutdown`, `idle` |
| `spawn` | Process spawning | `spawn [backend] [command]` (defaults: tmux, claude) |
| `identity` | Identity management | `show`, `set` |

## JSON Output

All commands support `--json` for machine-readable output. Place the flag before the subcommand:

```bash
clawteam --json team discover
clawteam --json board show my-team
clawteam --json task list my-team --status pending
```

Combine with `jq` for scripting:

```bash
clawteam --json board show my-team | jq '.taskSummary'
clawteam --json task list my-team | jq '.[].subject'
```

## Important Notes

- `inbox receive` **consumes** messages (deletes files). Use `inbox peek` for non-destructive reads.
- Task status `blocked` is **auto-set** when `--blocked-by` is specified at creation.
- Completing a task **auto-unblocks** any tasks that list it in `blockedBy`.
- `clawteam spawn` defaults to **tmux** backend with **git worktree** isolation and **skip-permissions**.
- All file writes use atomic tmp+rename to prevent data corruption.
- Identity env vars are set automatically when spawning via `clawteam spawn`.
- Use `board attach <team>` to watch all agents in a tiled tmux layout.

## Golang Development Skills Integration

When coordinating teams working on Go codebases, instruct spawned agents to use the appropriate
golang skills via `/skill-name` slash commands. This ensures agents follow idiomatic Go patterns,
use production-grade libraries, and maintain code quality across the team.

### Skill Selection by Task Type

When creating tasks or sending instructions to agents, reference the relevant skills so agents
invoke them. Include the skill slash command in the agent's task description or inbox message.

#### Code Architecture & Design
| Skill | Use When |
|-------|----------|
| `/golang-project-layout` | Setting up new Go projects, organizing monorepos, structuring packages |
| `/golang-design-patterns` | Functional options, constructors, error flow, graceful shutdown, DI |
| `/golang-structs-interfaces` | Designing structs, interfaces, embedding, type assertions, receivers |
| `/golang-naming` | Naming packages, functions, variables, interfaces, constants, enums |
| `/golang-code-style` | Code formatting, conventions, linter configuration, comment style |
| `/golang-modernize` | Upgrading code to use latest Go features and idioms |

#### Error Handling & Safety
| Skill | Use When |
|-------|----------|
| `/golang-error-handling` | Error creation, wrapping, Is/As, sentinel errors, panic/recover |
| `/golang-samber-oops` | Structured errors with samber/oops: stack traces, error codes, attributes |
| `/golang-safety` | Preventing panics, nil safety, numeric conversions, resource lifecycle |
| `/golang-security` | Injection prevention, cryptography, filesystem safety, secrets management |

#### Concurrency & Performance
| Skill | Use When |
|-------|----------|
| `/golang-concurrency` | Goroutines, channels, select, sync primitives, worker pools, fan-out |
| `/golang-context` | context.Context creation, propagation, cancellation, timeouts, values |
| `/golang-performance` | Allocation reduction, CPU efficiency, GC tuning, pooling, caching |
| `/golang-benchmark` | Writing benchmarks, profiling with pprof, benchstat analysis |

#### Testing & Quality
| Skill | Use When |
|-------|----------|
| `/golang-testing` | Table-driven tests, integration tests, fuzzing, parallel tests, fixtures |
| `/golang-stretchr-testify` | assert/require/mock/suite packages from testify |
| `/golang-lint` | golangci-lint configuration, nolint directives, linter settings |
| `/golang-troubleshooting` | Debugging crashes, deadlocks, unexpected behavior, pprof |

#### Data & Storage
| Skill | Use When |
|-------|----------|
| `/golang-data-structures` | Slices, maps, arrays, container packages, strings.Builder, generics |
| `/golang-database` | SQL queries, transactions, connection pools, migrations, scanning |

#### Networking & APIs
| Skill | Use When |
|-------|----------|
| `/golang-grpc` | gRPC servers/clients, protobuf, interceptors, streaming |
| `/golang-cli` | CLI apps: command structure, flags, config, signals, shell completion |

#### Observability & Operations
| Skill | Use When |
|-------|----------|
| `/golang-observability` | Structured logging, Prometheus metrics, OpenTelemetry tracing, pprof |
| `/golang-continuous-integration` | GitHub Actions, testing pipelines, SAST, GoReleaser, code coverage |
| `/golang-dependency-management` | go.mod management, versioning, vulnerability scanning, upgrades |
| `/golang-dependency-injection` | DI patterns, google/wire, uber-go/dig, samber/do comparison |

#### Samber Ecosystem Libraries
| Skill | Use When |
|-------|----------|
| `/golang-samber-lo` | Functional helpers: Map, Filter, Reduce, GroupBy, Chunk, Find, Uniq |
| `/golang-samber-mo` | Monadic types: Option, Result, Either, Future, IO, Task |
| `/golang-samber-do` | Dependency injection containers, service lifecycles |
| `/golang-samber-slog` | Logging extensions: multi-handler, sampling, HTTP middleware |
| `/golang-samber-hot` | In-memory caching: LRU, LFU, TinyLFU, TTL, sharding |
| `/golang-samber-ro` | Reactive streams: observables, subjects, operators, pipelines |
| `/golang-samber-oops` | Structured error handling with stack traces and context |

#### Documentation & Libraries
| Skill | Use When |
|-------|----------|
| `/golang-documentation` | Godoc comments, README, CONTRIBUTING, example tests, llms.txt |
| `/golang-popular-libraries` | Choosing production-ready libraries, comparing alternatives |
| `/golang-stay-updated` | Go news, communities, learning resources |

### Example: Spawning a Go Development Team

```bash
# Create a Go microservice development team
clawteam team spawn-team go-svc -d "Go microservice development" -n leader

# Spawn specialized Go agents with skill-aware task descriptions
clawteam spawn --team go-svc --agent-name architect \
  --task "Design the service architecture. Use /golang-project-layout for structure, \
/golang-design-patterns for patterns, and /golang-structs-interfaces for type design."

clawteam spawn --team go-svc --agent-name backend-dev \
  --task "Implement the API server. Use /golang-grpc for gRPC setup, \
/golang-error-handling for error patterns, /golang-concurrency for concurrent request handling, \
and /golang-context for context propagation."

clawteam spawn --team go-svc --agent-name test-dev \
  --task "Write comprehensive tests. Use /golang-testing for test patterns, \
/golang-stretchr-testify for assertions and mocks, /golang-benchmark for benchmarks, \
and /golang-lint for linter configuration."

clawteam spawn --team go-svc --agent-name ops-dev \
  --task "Set up observability and CI/CD. Use /golang-observability for logging/metrics/tracing, \
/golang-continuous-integration for GitHub Actions pipelines, \
and /golang-dependency-management for go.mod management."

# Send skill-specific guidance via inbox
clawteam inbox send go-svc backend-dev \
  "Before implementing, run /golang-security to review security best practices for the auth module. \
Also use /golang-samber-oops for structured error handling throughout."

clawteam inbox send go-svc architect \
  "Check /golang-popular-libraries before choosing any third-party dependencies. \
Use /golang-naming conventions for all exported identifiers."
```

### Skill Selection Guidelines for Leaders

When assigning Go tasks, the leader agent should:

1. **Match skills to task scope** — Include 2-4 relevant skill references per task. Don't overload agents with all skills.
2. **Pair implementation with quality skills** — Every coding task should reference at least one of: `/golang-testing`, `/golang-lint`, `/golang-safety`, or `/golang-security`.
3. **Use samber skills when the project uses samber libraries** — Check `go.mod` for `github.com/samber/*` imports before recommending samber-specific skills.
4. **Reference `/golang-code-style` and `/golang-naming` for consistency** — Especially when multiple agents write code that must integrate.
5. **Include `/golang-modernize`** when working on existing codebases to ensure agents use current Go idioms.

## Additional Resources

### Reference Files

For detailed command arguments, data models, and storage layout:
- **`references/cli-reference.md`** — Complete CLI reference with all commands, options, and data models

For step-by-step coordination workflows and common patterns:
- **`references/workflows.md`** — Multi-agent workflows: team setup, spawn coordination, join protocol, plan approval, graceful shutdown, monitoring patterns

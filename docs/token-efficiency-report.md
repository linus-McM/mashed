# Token Efficiency Report — mashed (Round 2)

**Date:** 2026-04-15
**Scope:** Post-round-1 audit. Prior fixes already landed: context-mode / expo / chrome-devtools disabled, 14 Sentry SDKs overridden, OPENWOLF.md trimmed to 67 lines.

## TL;DR

Baseline is now **~3–4K tokens/turn** (down from ~7–8K pre-round-1). Remaining bloat is smaller but still actionable: ~800–1,200 tokens/turn recoverable by disabling irrelevant plugins (cli-anything, sentry) and cleaning dead marketplace directories.

## Measurement Table (per-turn, auto-loaded only)

| Surface | Lines | ≈Tokens | Auto/Lazy |
|---|---|---|---|
| `CLAUDE.md` + `@.wolf/OPENWOLF.md` + `.claude/rules/openwolf.md` | 105 | ~1,260 | auto |
| `MEMORY.md` index | 2 | ~40 | auto |
| MCP instruction blocks (claude-in-chrome, Context7) | ~15 | ~180 | auto |
| Skill catalog (project 41 + plugin 115 = 156 names) | ~156 | ~2,000 | auto |
| SessionStart hooks | 0 entries | 0 | — |
| PreToolUse hooks | 0 entries | 0 | — |
| PostToolUse hooks (auto-commit, silent on success) | 2 entries | 0 per-turn | per-invoke only |
| Deferred tools list (names only) | ~60 | ~500 | auto |
| **Total recurring** | — | **~3,980** | |

## Findings

### HIGH — `cli-anything@cli-anything` plugin is irrelevant to this repo

Marketplace contents: OBS Studio, Audacity, Inkscape, Shotcut, codex-skill harness. Desktop-media automation plugins. Zero overlap with a Wails Go+Svelte IDE.

**Fix:** disable in `enabledPlugins` → removes plugin skill names from catalog (~80–120 tokens/turn).

### HIGH — `sentry@claude-plugins-official` plugin: 20 SDK skills still active

Mashed is a local desktop IDE with no Sentry integration (no `sentry/*` import, no DSN in config). Despite 14 SDK overrides landed in round 1, 20 Sentry skills remain in the catalog:

```
sentry-cloudflare-sdk, sentry-cocoa-sdk, sentry-code-review, sentry-create-alert,
sentry-dotnet-sdk, sentry-elixir-sdk, sentry-feature-setup, sentry-fix-issues,
sentry-go-sdk, sentry-nextjs-sdk, sentry-otel-exporter-setup, sentry-php-sdk,
sentry-pr-code-review, sentry-react-sdk, sentry-sdk-setup, sentry-sdk-skill-creator,
sentry-sdk-upgrade, sentry-setup-ai-monitoring, sentry-svelte-sdk, sentry-workflow
```

Plus Sentry MCP `authenticate` / `complete_authentication` deferred tools.

**Fix:** disable entire `sentry@claude-plugins-official` plugin. Saves ~500 tokens/turn. Also makes the 14 existing `sentry:*` skillOverrides redundant (can be removed).

### MED — `atomic-agents@claude-plugins-official` plugin: 7 skills, unclear usage

Contributes: atomic-agents, atomic-structure, atomic-prompts, atomic-schemas, atomic-context, atomic-tools, release. No references in repo code or .wolf/ artifacts. Ask user.

**Fix (if unused):** disable → ~150 tokens/turn.

### MED — Dead marketplace directories on disk

`~/.claude/plugins/marketplaces/context-mode/` and `~/.claude/plugins/marketplaces/expo-plugins/` remain after plugin disable. Don't add per-turn cost but consume disk and slow marketplace scans.

**Fix:** `rm -rf` both. Safe — plugins already disabled in `enabledPlugins`.

### LOW — Duplicated OpenWolf rules

`CLAUDE.md` + `.wolf/OPENWOLF.md` + `.claude/rules/openwolf.md` cover overlapping ground. The 15-line `rules/openwolf.md` restates OPENWOLF.md bullets.

**Fix:** delete `.claude/rules/openwolf.md` (covered by OPENWOLF.md via CLAUDE.md's `@` reference). Saves ~180 tokens/turn.

### LOW — 14 `sentry:*` skillOverrides become redundant if plugin disabled

If sentry plugin disabled, clean `skillOverrides` entries for `sentry:*` → settings.json slightly smaller, no runtime effect.

## Recommendations (priority order)

| # | Action | Savings/turn |
|---|---|---|
| 1 | Disable `sentry@claude-plugins-official` | ~500 |
| 2 | Disable `cli-anything@cli-anything` | ~100 |
| 3 | Delete `.claude/rules/openwolf.md` (duplicate) | ~180 |
| 4 | Disable `atomic-agents@claude-plugins-official` (if unused) | ~150 |
| 5 | `rm -rf` dead marketplace dirs (context-mode, expo-plugins) | 0 (disk only) |
| 6 | Clean redundant `sentry:*` overrides | 0 (cleanup) |
| **Total** | | **~930 tokens/turn** |

## Caveats

- Prompt cache (5-min TTL) amortizes most per-turn cost after the first message.
- Token estimates ±20%.
- Plugin disable requires **session restart** to take effect.
- Do NOT disable `sentry` plugin if you use Seer or Sentry MCP auth on other projects — this is a user-scope (global) setting. If you do, skip #1 and keep the SDK overrides.
- Keep: `gopls-lsp` (active Go dev), `frontend-design`, `commit-commands`, `caveman`, `claude-md-management`, `skill-creator`, `cc-caffeine`, `coderabbit`.

## How to apply

Run `/token-audit --apply` and multi-select from the prompt, or edit `~/.claude/settings.json` directly.

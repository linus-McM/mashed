# Design System — Claude Conductor

## Product Context
- **What this is:** Notification-first IDE for multi-agent development
- **Who it's for:** Solo AI-heavy developers running multiple Claude Code sessions
- **Space/industry:** Developer tools, AI agent orchestration
- **Project type:** Desktop app (Wails v2 + Svelte)

## Aesthetic Direction
- **Direction:** Industrial/Utilitarian
- **Decoration level:** Minimal — typography, color, and density do all the work
- **Mood:** Calm command center. Linear meets Bloomberg Terminal. Every pixel earns its place.
- **Reference sites:** Linear (surface hierarchy), Warp (accent color system), Zed (font rendering)

## Typography
- **Display/UI:** Geist (Vercel) — geometric, clean, modern. Distinct from Inter.
- **Body:** Geist — same family for consistency
- **Mono/Data:** Geist Mono — tabular numbers, sparklines, token counts
- **Code/Terminal:** JetBrains Mono — developer standard for terminal panes
- **Loading:** Google Fonts CDN or self-hosted via Wails embedded assets
- **Scale:** 11px (labels) / 13px (body) / 14px (data) / 16px (section headers) / 20px (page titles) / 28px (hero)

## Color
- **Approach:** Restrained — one neon accent, semantic status colors, muted neutrals
- **Backgrounds:** #07080a (deepest) → #0d0f12 (surface) → #12151a (elevated) → #181c23 (active)
- **Borders:** #1e2530 (subtle) / #2a3340 (emphasized)
- **Primary accent:** #00e57a (neon green) — "alive, active, go." The brand color.
- **Accent dim:** #006636 (muted green for backgrounds)
- **Alert:** #f0a500 (amber) — needs human attention
- **Error:** #e84545 (red) — something failed
- **Info:** #3d9eff (blue) — informational
- **Agent/System:** #9d6fff (purple) — sub-agent/skill events
- **Teal:** #00c4b3 — secondary accent (sparklines, progress)
- **Text:** #c8d4e0 (primary) / #4a5a6a (dim) / #2e3d4d (muted)
- **Dark mode:** This IS dark mode. No light mode in V1.

## Spacing
- **Base unit:** 4px
- **Density:** Compact (dev tools need density, not whitespace)
- **Scale:** 2xs(2px) xs(4px) sm(8px) md(12px) lg(16px) xl(24px) 2xl(32px) 3xl(48px)

## Layout
- **Approach:** App UI — primary workspace dominates, thin chrome
- **Grid:** Single-column feed (V1), split-pane for agent detail
- **Max content width:** None (full window)
- **Border radius:** sm:2px (subtle), md:4px (cards/inputs), lg:8px (modals), none for terminal panes

## Motion
- **Approach:** Minimal-functional
- **Notification entry:** 150ms ease-out (slide + fade from top)
- **State transitions:** 100ms color transition
- **Terminal:** Zero motion (instant rendering)
- **Easing:** enter(ease-out) exit(ease-in) move(ease-in-out)
- **Duration:** micro(50ms) short(100ms) medium(150ms) — nothing longer

## Component Patterns
- **Notification row:** Left accent stripe (4px, status color) + repo + agent + summary + time
- **Status badge:** Pill shape, background: status color at 15% opacity, text: status color
- **Sparkline:** Geist Mono, block characters (▁▂▃▄▅▆▇█), teal color
- **Terminal pane:** No border-radius, #07080a background, JetBrains Mono 13px
- **Diff view:** Green(+) / Red(-) with line numbers in dim text
- **Worktree panel:** Border-top separator, compact layout, action buttons as text links

## Risks (deliberate departures)
- **Neon green accent** — louder than typical dev tools. Intentional: "alive" should feel electric.
- **Geist typography** — newer, less battle-tested than Inter. Gain: distinctive geometric feel.
- **Zero window chrome** — notification feed IS the whole screen. Gain: maximum content density.

## Decisions Log
| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-04-01 | Initial design system | Created by /design-consultation. Industrial/Utilitarian aesthetic, neon green accent, Geist typography. |

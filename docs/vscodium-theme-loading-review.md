# Adversarial Review: VSCodium Theme Loading Plan

**Spec Location:** `/Users/linus/Development/mashed/docs/vscodium-theme-loading-plan.md`
**Review Date:** 2026-04-03
**Total Phases:** 6
**Total Approach Tiers:** 3 (A, B, C)

## Review Summary

- **CRITICAL Issues:** 5 (will cause build failures, panics, or broken behavior)
- **HIGH Issues:** 7 (will cause incorrect behavior or require rework)
- **MEDIUM Issues:** 6 (may cause confusion or require rework)
- **LOW Issues:** 4 (style issues, minor inconsistencies)

---

## CRITICAL Issues

### [C-1] Path Traversal Bypass via Symlinks in ReadThemeFile Security Check

**Spec Location:** Phase 1, Section 1.3 "ReadThemeFile"

**What Spec Claims:**
> Security: verify the path is within the configured extensions directory

The security check uses `filepath.Abs()` + `strings.HasPrefix()`:
```go
absTheme, _ := filepath.Abs(themePath)
absExt, _ := filepath.Abs(extDir)
if !strings.HasPrefix(absTheme, absExt) {
    return "", fmt.Errorf("theme path is outside extensions directory")
}
```

**Actual Problem:**
1. `filepath.Abs` does NOT resolve symlinks. An attacker (or a malicious extension) could create a symlink inside the extensions directory pointing to `/etc/passwd` or `~/.ssh/id_rsa`. `filepath.Abs` would make it look like it's inside `extDir`, but `os.ReadFile` follows the symlink.
2. The errors from `filepath.Abs` and `os.UserHomeDir` are silently discarded with `_ =`. If `UserHomeDir` fails, `extDir` becomes `/extDir[1:]` (garbage), and the prefix check is meaningless.
3. The check reads the file at `themePath` (the user-supplied argument), NOT at `absTheme`. If `themePath` is `../../../etc/passwd` and `extDir` is empty/broken, the prefix check passes vacuously.

**Impact:** Path traversal vulnerability. A crafted `themePath` string from the frontend could read arbitrary files on the host filesystem.

**Recommended Fix:**
```go
absTheme, err := filepath.EvalSymlinks(themePath)
if err != nil {
    return "", fmt.Errorf("resolving theme path: %w", err)
}
absExt, err := filepath.EvalSymlinks(extDir)
if err != nil {
    return "", fmt.Errorf("resolving extensions path: %w", err)
}
if !strings.HasPrefix(absTheme+string(os.PathSeparator), absExt+string(os.PathSeparator)) {
    return "", fmt.Errorf("theme path is outside extensions directory")
}
```
Also add `string(os.PathSeparator)` suffix to prevent prefix confusion (e.g., `/ext-other/` matching `/ext/`).

---

### [C-2] Store Refactor Breaks TitleBar.svelte (and Potentially Other Consumers)

**Spec Location:** Phase 3, "Modified `stores/theme.js`"

**What Spec Claims:**
The spec proposes changing `stores/theme.js` to export `themeIds` as a `derived` store:
```js
export const themeIds = derived(allThemes, ($all) => Object.keys($all));
```
And `themes` as a backward-compat alias:
```js
export const themes = builtInThemes;
```

**Actual Codebase State:**
- File: `/Users/linus/Development/mashed/frontend/src/components/TitleBar.svelte:5`
  ```js
  import { themes, themeIds, currentThemeId, applyTheme } from '../lib/stores/theme.js';
  ```
- File: `/Users/linus/Development/mashed/frontend/src/components/TitleBar.svelte:50-51`
  ```svelte
  {#each themeIds as id}
    {@const theme = themes[id]}
  ```

Currently `themeIds` is a plain array (`Object.keys(themes)`). After the refactor, `themeIds` becomes a Svelte derived store. Using `{#each themeIds as id}` on a store (not `$themeIds`) will iterate over the store object's properties, NOT the array values. This will silently produce zero iterations or garbage.

Additionally, `themes` is changed from the full themes map to just `builtInThemes` -- so `themes[id]` will return `undefined` for any imported theme selected from the TitleBar picker.

**Impact:** The TitleBar theme picker will render nothing or crash after the store refactor. This is a silent regression.

**Files Affected:**
- `/Users/linus/Development/mashed/frontend/src/components/TitleBar.svelte:5,50-51`
- `/Users/linus/Development/mashed/frontend/src/views/Settings.svelte:4,67-69`

**Recommended Fix:**
1. TitleBar.svelte must be updated to use `$themeIds` and `$allThemes[$currentThemeId]` instead of `themes[id]`.
2. Settings.svelte's existing theme grid must also switch to `$allThemes`.
3. The spec should explicitly list TitleBar.svelte in the "File Changes Summary" table -- it is currently omitted.

---

### [C-3] JSONC Theme Files Will Fail with JSON.parse

**Spec Location:** Phase 4, `activateImportedTheme` function

**What Spec Claims:**
```js
const raw = await ReadThemeFile(themePath);
const vsTheme = JSON.parse(raw);
```

**Actual Problem:**
Many VSCode theme files use JSONC (JSON with Comments) format. This is explicitly allowed by the VSCode extension API. Popular themes like One Dark Pro, Dracula, and GitHub use `//` comments and trailing commas in their theme JSON files. `JSON.parse()` will throw a SyntaxError on any such file.

The spec mentions `.tmTheme` filtering (line 295) but never mentions JSONC at all -- the word "comment" only appears in the context of syntax highlighting scopes, never in the context of the theme file format itself.

**Impact:** A significant percentage of real-world VSCode themes will fail to load with an opaque "Failed to load theme" error.

**Recommended Fix:**
Either:
1. Strip comments in the Go backend before returning the string (using a simple regex or a JSONC-aware parser), or
2. Use a JSONC parser on the frontend: `npm install jsonc-parser` and replace `JSON.parse(raw)` with `jsonc.parse(raw)`, or
3. Strip comments with a simple JS function before parsing:
```js
function stripJsonComments(str) {
  return str.replace(/\/\/.*$/gm, '').replace(/\/\*[\s\S]*?\*\//g, '');
}
```
Option 1 (Go backend) is cleanest since it centralizes the parsing concern.

---

### [C-4] Size Check After Full File Read in ReadThemeFile

**Spec Location:** Phase 1, Section 1.3

**What Spec Claims:**
```go
data, err := os.ReadFile(themePath)
if err != nil {
    return "", fmt.Errorf("reading theme file: %w", err)
}
// Cap at 512KB
if len(data) > 512*1024 {
    return "", fmt.Errorf("theme file too large (>512KB)")
}
```

**Actual Problem:**
The code reads the ENTIRE file into memory, THEN checks the size. If a symlink or adversarial file points to a multi-gigabyte file, the Go process will allocate all that memory before rejecting it. This is a denial-of-service vector.

**Impact:** Memory exhaustion if a large file is encountered (even accidentally -- e.g., a binary blob in an extension directory).

**Recommended Fix:**
Use `os.Stat` to check size BEFORE reading:
```go
info, err := os.Stat(themePath)
if err != nil {
    return "", fmt.Errorf("stat theme file: %w", err)
}
if info.Size() > 512*1024 {
    return "", fmt.Errorf("theme file too large (%d bytes, max 512KB)", info.Size())
}
data, err := os.ReadFile(themePath)
```

---

### [C-5] App.svelte Theme Restore Does Not Handle Imported Themes on Startup

**Spec Location:** Phase 4, Settings.svelte `onMount`

**What Spec Claims:**
Settings.svelte handles re-applying imported themes in its `onMount`:
```js
if (cfg.importedTheme) {
    await activateImportedTheme(cfg.importedTheme);
}
```

**Actual Codebase State:**
- File: `/Users/linus/Development/mashed/frontend/src/App.svelte:22-23`
  ```js
  const cfg = await GetConfig();
  if (cfg.theme) applyTheme(cfg.theme);
  ```

App.svelte runs on every app startup and tries to apply the saved theme ID (e.g., `"imported-dracula"`). But at startup, the `allThemes` store only contains built-in themes. The imported theme has not been loaded/converted yet -- that only happens when the user navigates to Settings. So `applyTheme("imported-dracula")` will silently fail (the theme won't exist in the store), and the app will render with no theme applied (raw CSS defaults).

**Impact:** Every app restart loses the imported theme until the user manually navigates to Settings. The theme flickers/resets on every launch.

**Recommended Fix:**
The imported theme restoration logic must run in `App.svelte`'s `onMount`, not just in Settings.svelte. Either:
1. Move the `activateImportedTheme` function to a shared module and call it from App.svelte, or
2. Create an `initThemeFromConfig()` function in `stores/theme.js` that handles both built-in and imported theme restoration.

---

## HIGH Issues

### [H-1] `ReadThemeFileResolved` Introduced in Phase 6 but Never Referenced in Phase 4

**Spec Location:** Phase 6, Section 6.1 vs. Phase 4

**What Spec Claims:**
Phase 6 adds `ReadThemeFileResolved()` to handle theme inheritance (`"include"` field). But Phase 4's `activateImportedTheme` calls `ReadThemeFile()`, not `ReadThemeFileResolved()`.

**Impact:** Themes with `include` directives (common in theme families like Catppuccin, which shares a base) will load with missing colors and token rules. The "include" resolution is implemented but never actually wired up.

**Recommended Fix:**
Phase 4 should call `ReadThemeFileResolved` instead of `ReadThemeFile`, or `ReadThemeFile` itself should be modified to resolve includes (rather than creating a separate function).

---

### [H-2] Race Condition in `loadConfig()` / `saveConfig()` — No File Locking

**Spec Location:** Phase 1, Section 1.4

**What Spec Claims:**
`SetImportedTheme`, `SetTheme`, and `SetVSCodiumExtPath` all follow the load-modify-save pattern:
```go
cfg := loadConfig()
cfg.ImportedTheme = themePath
return saveConfig(cfg)
```

**Actual Codebase State:**
- File: `/Users/linus/Development/mashed/app.go:57-78`
- Multiple Wails-bound methods use this pattern concurrently

Since Wails dispatches method calls from the frontend JavaScript, rapid theme switching or simultaneous config changes can cause a TOCTOU race: two goroutines read the same config, each modifies a different field, and the last writer wins (clobbering the other's change).

**Impact:** Config corruption under concurrent access. Example: user changes VSCodiumExtPath while the theme is being saved -- one write is lost.

**Recommended Fix:**
Add a mutex around all config read-modify-write operations, or use the existing `a.mu` mutex for config access:
```go
func (a *App) SetImportedTheme(themePath string) error {
    a.mu.Lock()
    defer a.mu.Unlock()
    cfg := loadConfig()
    cfg.ImportedTheme = themePath
    return saveConfig(cfg)
}
```

---

### [H-3] `~` Expansion Bug: `~/` vs `~` Prefix

**Spec Location:** Phase 1, Section 1.2

**What Spec Claims:**
```go
if strings.HasPrefix(extDir, "~") {
    home, _ := os.UserHomeDir()
    extDir = filepath.Join(home, extDir[1:])
}
```

**Actual Problem:**
This matches ANY string starting with `~`, not just `~/`. A path like `~otheruser/extensions` would be incorrectly transformed to `/Users/linus/otheruser/extensions` instead of being treated as a different user's home directory (Unix convention).

More critically, `extDir[1:]` when `extDir` is `~/foo` produces `/foo`, so `filepath.Join(home, "/foo")` works correctly. But if `extDir` is just `~`, `extDir[1:]` is `""`, which is fine. However, the `os.UserHomeDir()` error is again silently discarded.

**Impact:** Incorrect path resolution for `~otheruser/` paths (edge case). Silent failures if home dir lookup fails.

**Recommended Fix:**
```go
if strings.HasPrefix(extDir, "~/") || extDir == "~" {
    home, err := os.UserHomeDir()
    if err != nil {
        return nil, fmt.Errorf("resolving home directory: %w", err)
    }
    extDir = filepath.Join(home, extDir[2:])
}
```

---

### [H-4] `applyTheme()` in Proposed Store Only Updates CSS, Not xterm

**Spec Location:** Phase 3, "Modified `stores/theme.js`"

**What Spec Claims:**
```js
export function applyTheme(id) {
  const all = get(allThemes);
  const theme = all[id];
  if (!theme) return;
  for (const [prop, value] of Object.entries(theme.css)) {
    document.documentElement.style.setProperty(prop, value);
  }
  currentThemeId.set(id);
}
```

**Actual Codebase State:**
Terminal.svelte reactively updates xterm via:
```js
$: if (term && $currentTheme) {
    term.options.theme = $currentTheme.xterm;
}
```
This depends on `currentTheme` being derived from `currentThemeId + allThemes`. The proposed store DOES derive `currentTheme` correctly, so this part is fine.

However, the `applyTheme` function itself does not differ from the current one functionally -- but the current `applyTheme` looks up from the static `themes` object, while the proposed one uses `get(allThemes)`. The issue is that `applyTheme` is called from App.svelte on startup BEFORE any imported theme is registered in `allThemes`. This connects to C-5.

**Impact:** Startup theme application will silently fail for imported themes.

---

### [H-5] `convertTokenColors` Uses First-Match-Wins, Losing Specificity

**Spec Location:** Phase 2, Approach B2, `convertTokenColors` function

**What Spec Claims:**
```js
if (match && !seenTokens.has(mapping.token)) {
    // ...
    seenTokens.add(mapping.token);
    break;
}
```

**Actual Problem:**
The `seenTokens` set prevents a Monaco token type from being assigned more than once. But VSCode themes intentionally define multiple rules with increasing specificity. For example:
```json
{ "scope": "comment", "settings": { "foreground": "#6272A4" } },
{ "scope": "comment.block.documentation", "settings": { "foreground": "#8899AA", "fontStyle": "italic" } }
```
The first `comment` scope would map to Monaco's `comment` token and mark it as "seen". The more specific `comment.block.documentation` scope would be skipped entirely because `comment` is already in `seenTokens`.

Since VSCode themes list scopes from general to specific, the FIRST (least specific) match wins. This is backwards -- the most specific should win.

**Impact:** Token colors will be less accurate than expected for themes with detailed scope hierarchies.

**Recommended Fix:**
Process tokenColors in reverse order (most specific first), OR don't use `seenTokens` at all and let Monaco's own specificity handling take care of conflicts (Monaco applies the last matching rule).

---

### [H-6] Monaco `colors` Object Receives Raw VSCode Color Strings Including Alpha

**Spec Location:** Phase 2, all approaches

**What Spec Claims:**
```js
colors: { ...colors },  // Pass through all VSCode editor color keys
```

**Actual Problem:**
VSCode color values can include alpha channels in the form `#RRGGBBAA` (8-character hex). For example, `"editor.selectionBackground": "#44475a75"`. Monaco's `defineTheme` expects colors in `#RRGGBB` or `#RRGGBBAA` format -- this is actually fine for the `colors` object.

However, the CSS variable mapper (`mapVSCodeColorsToCSSVars`) passes these 8-char hex values directly into CSS custom properties. CSS supports `#RRGGBBAA` in modern browsers, but some of these values are used in contexts where alpha might produce unexpected results (e.g., `--bg-deepest` with alpha would show through to the window background).

**Impact:** Some themes with transparent backgrounds or selections may render with unexpected see-through regions.

**Recommended Fix:**
Add a note in the spec to strip alpha from background color values, or convert to `rgba()` format for CSS.

---

### [H-7] `.tmTheme` Filter Missing from Scanner Code

**Spec Location:** Phase 1, Section 1.2 + Section 1.6

**What Spec Claims (Section 1.6):**
> The scanner should filter for `.json` files only — `.tmTheme` conversion requires additional parsing and is not worth the complexity for the initial implementation.

**What the Code Actually Does (Section 1.2):**
```go
themePath := filepath.Join(extDir, entry.Name(), t.Path)
if _, err := os.Stat(themePath); err != nil {
    continue
}
```
The scanner checks if the file exists but does NOT filter by extension. If a `package.json` declares a theme with `"path": "./themes/Monokai.tmTheme"`, the scanner will include it in the results, and the frontend will fail when trying to `JSON.parse()` an XML file.

**Impact:** `.tmTheme` themes will appear in the picker but crash when selected.

**Recommended Fix:**
Add an extension check:
```go
if !strings.HasSuffix(strings.ToLower(t.Path), ".json") {
    continue
}
```

---

## MEDIUM Issues

### [M-1] No Concurrency Protection for `importedThemeMap` in Settings.svelte

**Spec Location:** Phase 4, `activateImportedTheme`

The `importedThemeMap` object is read and written in an async function. If the user clicks multiple theme buttons rapidly, multiple concurrent invocations of `activateImportedTheme` could race on the map check and trigger redundant `ReadThemeFile` calls. While this won't corrupt data (JS is single-threaded for sync code), the async `ReadThemeFile` await creates an interleaving window where the wrong theme could be applied last.

**Recommended Fix:**
Add a guard variable:
```js
let activatingTheme = null;
async function activateImportedTheme(themePath) {
    if (activatingTheme === themePath) return;
    activatingTheme = themePath;
    try { /* ... */ } finally { activatingTheme = null; }
}
```

---

### [M-2] Spec Claims `wails generate module` -- Command Does Not Exist

**Spec Location:** Phase 1, Section 1.5

**What Spec Claims:**
```bash
wails generate module
```

**Actual Problem:**
The correct Wails v2 command is `wails generate` (no `module` subcommand). Alternatively, running `wails dev` or `wails build` auto-generates bindings. The command `wails generate module` will fail with an unknown subcommand error.

**Recommended Fix:**
Replace with `wails generate` or just note that `wails dev` will auto-generate bindings.

---

### [M-3] `conductorConfig` Struct Redefined in Spec Without Acknowledging Existing Fields

**Spec Location:** Phase 1, Section 1.4

**What Spec Claims:**
```go
type conductorConfig struct {
    DevDir          string `json:"devDir"`
    Theme           string `json:"theme,omitempty"`
    VSCodiumExtPath string `json:"vscodiumExtPath,omitempty"`
    ImportedTheme   string `json:"importedTheme,omitempty"` // NEW
}
```

**Actual Codebase State:**
- File: `/Users/linus/Development/mashed/app.go:44-48`

The existing struct is identical to the spec's version minus `ImportedTheme`. This is fine, but the spec presents the FULL struct as if it needs to be replaced, rather than showing just the field addition. A literal copy-paste would be harmless, but the pattern suggests the spec author may not have verified the current state.

**Recommended Fix:**
Clarify: "Add the following field to the existing `conductorConfig`" rather than showing the full struct.

---

### [M-4] Approach B1 References npm Package `monaco-vscode-textmate-theme-converter` — May Not Exist

**Spec Location:** Phase 2, Approach B1

**What Spec Claims:**
```bash
npm install monaco-vscode-textmate-theme-converter
```

**Actual Problem:**
This package name is suspiciously specific and may not exist on npm. The spec notes it was "last updated in 2023" but does not provide a link or verify its existence. A search for packages with similar names reveals `monaco-vscode-textmate-theme-converter` is not a well-known package. The actual common package for this purpose is `@vscode/textmate` or manual conversion.

**Impact:** `npm install` will fail if the package doesn't exist.

**Recommended Fix:**
Verify the package exists on npm before recommending it. The spec already recommends B2 (hand-rolled) as the primary approach, so this is less critical -- but should be flagged or removed to avoid confusion.

---

### [M-5] Missing Theme ID Collision Handling

**Spec Location:** Phase 4, `activateImportedTheme`

**What Spec Claims:**
```js
const themeId = 'imported-' + themePath.split('/').pop().replace('.json', '');
```

**Actual Problem:**
If two extensions contain a file named `dark.json` (extremely common -- GitHub Theme, Material Theme, etc.), they'll both get the ID `imported-dark`. The second one will silently overwrite the first in the store.

**Impact:** Theme name collisions. Users cannot distinguish between two themes with the same filename.

**Recommended Fix:**
Include the extension ID in the theme ID:
```js
const themeId = 'imported-' + entry.extensionId + '-' + themePath.split('/').pop().replace('.json', '');
```

---

### [M-6] `ReadThemeFileResolved` Recursive Include Creates Infinite Loop Risk

**Spec Location:** Phase 6, Section 6.1

**What Spec Claims:**
```go
if theme.Include != "" {
    baseDir := filepath.Dir(themePath)
    includePath := filepath.Join(baseDir, theme.Include)
    baseData, err := a.ReadThemeFile(includePath)
```

**Actual Problem:**
If theme A includes theme B and theme B includes theme A, this creates an infinite recursion (via `ReadThemeFile` -> `ReadThemeFileResolved`). Actually, it calls `ReadThemeFile` (not `ReadThemeFileResolved`), so it only resolves one level. But this is itself a bug -- deeply nested includes (A includes B, B includes C) will only resolve one level.

**Impact:** Multi-level theme inheritance silently produces incomplete themes.

**Recommended Fix:**
Either:
1. Document the one-level limit explicitly, or
2. Use `ReadThemeFileResolved` recursively with a depth counter:
```go
func (a *App) readThemeResolved(themePath string, depth int) (string, error) {
    if depth > 5 { return "", fmt.Errorf("theme include depth exceeded") }
    // ...
}
```

---

## LOW Issues

### [L-1] `dimColor` Function Does Not Handle 4-char or 8-char Hex

**Spec Location:** Phase 2, `dimColor` helper

```js
function dimColor(hex, factor) {
  const r = parseInt(hex.slice(1, 3), 16);
```

VSCode colors can be `#RGB`, `#RRGGBB`, or `#RRGGBBAA`. The function assumes 7-character `#RRGGBB` format. A 4-char `#RGB` or 9-char `#RRGGBBAA` hex will produce wrong values.

**Recommended Fix:** Normalize hex input before parsing.

---

### [L-2] Decision Matrix Claims B2 Has "0" Dependencies but B2 Uses `dimColor` Utility

**Spec Location:** Decision Matrix table

This is not a dependency issue per se -- `dimColor` is defined inline. The "0" dependency claim is accurate. No action needed, but the matrix could note that B2 requires ~200 lines of converter code.

---

### [L-3] `hc-black` and `hc-light` UITheme Values Not Handled

**Spec Location:** Phase 2, all approaches

The spec checks `isDark = vsTheme.type !== 'light'` and defaults to `vs-dark`. But VSCode has four UI themes: `vs`, `vs-dark`, `hc-black`, `hc-light`. High-contrast themes should map to `hc-black` or `hc-light` base in Monaco, not `vs-dark`.

**Recommended Fix:**
```js
function getMonacoBase(vsTheme) {
    switch (vsTheme.type || vsTheme.uiTheme) {
        case 'light': case 'vs': return 'vs';
        case 'hc-black': return 'hc-black';
        case 'hc-light': return 'hc-light';
        default: return 'vs-dark';
    }
}
```

---

### [L-4] Spec Does Not Consider Very Large Extension Directories

**Spec Location:** Phase 1, Section 1.2

If the user points to a directory with hundreds of extensions (power users with 200+ extensions), `ListVSCodiumThemes` will synchronously read hundreds of `package.json` files. This could take several seconds and block the Wails call.

**Recommended Fix:** Note the potential for slowness and consider caching results or showing a progress indicator.

---

## Verification Checklist

- [x] All file references verified to exist (`themes.js`, `stores/theme.js`, `monacoTheme.js`, `Settings.svelte`, `MonacoEditor.svelte`)
- [x] All existing function/method names verified correct
- [x] All type/struct definitions checked against codebase (`conductorConfig` matches except new `ImportedTheme` field)
- [x] Go packages verified in go.mod (no new Go deps needed)
- [x] Import paths verified to resolve (Wails bindings exist at claimed path)
- [x] Wails binding compatibility verified (string/slice/struct return types are all supported)
- [x] Concurrency patterns reviewed (config race condition found)
- [x] Error handling patterns checked (silent `_ =` discards found)
- [x] Known patterns cross-referenced (error wrapping style matches project convention)
- [x] Frontend framework version verified (Svelte 3.59.2, Monaco 0.55.1)
- [x] All consumers of modified exports identified (TitleBar.svelte, Terminal.svelte, App.svelte)

## Positive Findings

1. **Architecture is sound.** The separation of Go backend (filesystem access) and Svelte frontend (conversion/registration) is well-motivated and plays to each layer's strengths.
2. **The three-tier approach is well-structured.** Each tier builds cleanly on the previous one, and the recommendation (B2) is pragmatic.
3. **The existing codebase is well-organized.** The `conductorConfig` already has `VSCodiumExtPath`, `themes.js` has a clean unified format, and `monacoTheme.js` is minimal. The plan extends these naturally.
4. **The xterm theme derivation is thorough.** Mapping ANSI colors from the VSCode terminal palette is correct and well-considered.
5. **The phase ordering allows incremental testing.** Each phase is independently verifiable, which is excellent for avoiding a big-bang integration.
6. **The spec correctly identifies that Monaco's `colors` keys match VSCode's** -- this is accurate and avoids unnecessary mapping work.

## Recommended Next Steps (Prioritized)

1. **Fix C-1 (path traversal)** — Use `filepath.EvalSymlinks` and add separator suffix to prefix check. Non-negotiable security fix.
2. **Fix C-3 (JSONC)** — Add JSONC stripping. Without this, a large percentage of real-world themes will fail.
3. **Fix C-5 (startup restore)** — Move imported theme restoration to App.svelte or a shared init function. Without this, the feature feels broken on every restart.
4. **Fix C-2 (TitleBar breakage)** — Update TitleBar.svelte and add it to the file changes list. Update Settings.svelte theme grid to use the new store shape.
5. **Fix C-4 (size check)** — Move `os.Stat` before `os.ReadFile`.
6. **Fix H-7 (.tmTheme filter)** — Add `.json` extension filter in the scanner.
7. **Fix H-1 (ReadThemeFileResolved wiring)** — Either wire it into the flow or merge the include resolution into ReadThemeFile.
8. **Fix H-5 (token specificity)** — Reverse iteration order or remove seenTokens set.
9. **Fix M-5 (ID collisions)** — Include extension ID in theme ID generation.
10. **Add JSONC to edge cases section** (Phase 6) as an explicit consideration.

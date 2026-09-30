// Estimates the cols/rows to spawn a PTY at so claude/zsh paint at roughly
// the right size on first frame. The Terminal component still sends a precise
// resize once xterm has measured its container, but the spawn-time estimate
// closes the visible gap where claude's banner would otherwise render at the
// helper default of 80x24 and end up wedged in scrollback at the wrong size.
//
// 9px char width / 18px line height match the xterm defaults at 15px
// monospace on macOS within a few pixels — close enough that the post-mount
// resize is invisible.
const APPROX_CHAR_WIDTH_PX = 9;
const APPROX_LINE_HEIGHT_PX = 18;

// Reserve room for the right-hand panel, header chrome, and tab strip.
const HORIZONTAL_CHROME_PX = 320;
const VERTICAL_CHROME_PX = 80;

export function estimatePtySize(): { cols: number; rows: number } {
  const w = typeof window !== 'undefined' ? window.innerWidth : 1280;
  const h = typeof window !== 'undefined' ? window.innerHeight : 800;
  const cols = Math.max(80, Math.floor((w - HORIZONTAL_CHROME_PX) / APPROX_CHAR_WIDTH_PX));
  const rows = Math.max(24, Math.floor((h - VERTICAL_CHROME_PX) / APPROX_LINE_HEIGHT_PX));
  return { cols, rows };
}

<script>
  // Tone-coloured banner — spec §6.2, Design Brief §3 + §5.
  //
  // Mirrors the `.prompt-block` left-stripe motif: 3px stripe in tone colour,
  // 10%-alpha tint background, radius flat on the stripe side. Each of the
  // four tones maps to one accent token — no mixing, no in-between shades.
  import { Info, AlertTriangle, XCircle, CheckCircle } from 'lucide-svelte';
  import { scrubUnsafeUrls } from './linkSanitiser';

  /** @type {'info'|'warn'|'error'|'success'} */
  export let tone = 'info';
  export let content = '';

  const ICON = {
    info: Info,
    warn: AlertTriangle,
    error: XCircle,
    success: CheckCircle,
  };

  $: Icon = ICON[tone] ?? Info;
  $: safeContent = scrubUnsafeUrls(content);
</script>

<div class="hint-banner" data-testid="hint-banner" data-tone={tone}>
  <span class="icon" data-testid="hint-banner-icon" aria-hidden="true">
    <svelte:component this={Icon} size={16} />
  </span>
  <span class="tone-label" data-testid="hint-banner-label">{tone.toUpperCase()}</span>
  <div class="body">{safeContent}</div>
</div>

<style>
  .hint-banner {
    display: flex;
    align-items: flex-start;
    gap: var(--sp-sm);
    padding: var(--sp-sm) var(--sp-md);
    border-radius: 0 var(--radius-md) var(--radius-md) 0;
    border-left: 3px solid var(--tone-color);
    background: color-mix(in srgb, var(--tone-color) 10%, transparent);
    color: var(--text-primary);
    font-family: var(--font-ui);
    font-size: var(--text-body);
    font-weight: 500;
    line-height: 1.5;
  }

  .hint-banner[data-tone='info']    { --tone-color: var(--accent-blue); }
  .hint-banner[data-tone='warn']    { --tone-color: var(--accent-amber); }
  .hint-banner[data-tone='error']   { --tone-color: var(--accent-red); }
  .hint-banner[data-tone='success'] { --tone-color: var(--accent-green); }

  .icon {
    color: var(--tone-color);
    display: inline-flex;
    flex: 0 0 auto;
    margin-top: 1px;
  }

  .tone-label {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    font-weight: 600;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    color: var(--tone-color);
    line-height: 1;
    flex: 0 0 auto;
    margin-top: 2px;
  }

  .body {
    flex: 1 1 auto;
    min-width: 0;
  }
</style>

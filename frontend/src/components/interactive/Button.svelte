<script lang="ts">
  import type { Snippet } from "svelte";
  import { ColorKeys } from "../../types/colors";
  import { addRipple, focusIndicator } from "../../lib/client/decoration";

  interface Props {
    onClick?: () => void;
    color?: ColorKeys;
    type?: "button" | "submit";
    compact?: boolean;
    plain?: boolean;
    enabled?: boolean;
    href?: string;
    title?: string;
    children?: Snippet;
  }

  let {
    onClick = () => {},
    color = ColorKeys.Neutral,
    type = "button",
    compact = false,
    plain = false,
    enabled = true,
    href = "",
    title = "",
    children
  }: Props = $props();
</script>

<style lang="scss">
  @use "sass:map";

  @use "../../styles/animations.scss";
  @use "../../styles/colors.scss";
  @use "../../styles/dimensions.scss";
  @use "../../styles/text.scss";

  button, a {
    // unset props
    background: none;
    color: inherit;
    border: none;
    padding: 0;
    font: inherit;
    cursor: pointer;
    outline: inherit;
    text-decoration: none;

    display: inline;

    cursor: pointer;
    padding: dimensions.$gapSmall;
    border-radius: dimensions.$borderRadius;

    min-width: dimensions.$buttonMinWidth;
    text-align: center;
    
    position: relative;
    overflow: hidden; 
  }

  // Nasty hack to get <Spinner/> to have the same height as text inside of buttons.
  button > :global(span.spinner) {
    &::before, &::after {
      content: "a";
      visibility: hidden;
    }
  }

  button:not(.neutral) {
    --barFocusIndicatorColor: #{colors.$barFocusIndicatorColorAlt};
  }

  button.compact, a.compact {
    min-width: dimensions.$buttonMinWidthCompact;
  }

  .disabled {
    cursor: not-allowed;
  }

  @each $key, $val in colors.$specialColors {
    button.#{$key}, a.#{$key} {
      background-color: map.get($val, "background");
      color: map.get($val, "foreground");
    }
    button.#{$key}.disabled, a.#{$key}.disabled {
      color: color-mix(in srgb, map.get($val, "foreground") 50%, transparent);
    }
  }

  button.plain, a.plain {
    background-color: transparent;
    color: colors.$foregroundPrimary;
    min-width: unset;
    padding: dimensions.$gapSmaller;
  }
  button.plain.disabled, a.plain.disabled {
    color: color-mix(in srgb, colors.$foregroundPrimary 50%, transparent);
  }
</style>

{#if href !== ""}
  <a
    class:success={!plain && color == ColorKeys.Success}
    class:warning={!plain && color == ColorKeys.Warning}
    class:danger={!plain && color == ColorKeys.Danger}
    class:accent={!plain && color == ColorKeys.Accent}
    class:neutral={!plain && color == ColorKeys.Neutral}
    class:inherit={!plain && color == ColorKeys.Inherit}
    class:compact={compact}
    class:plain={plain}
    onmouseleave={(e) => {(e.target as HTMLButtonElement).blur()}}
    class:disabled={!enabled}
    href={enabled ? href : "#"}
    title={title || undefined}
    aria-label={title || undefined}
    onmousedown={addRipple}
    use:focusIndicator
  >
    {@render children?.()}
  </a>
{:else}
  <button
    class:success={!plain && color == ColorKeys.Success}
    class:warning={!plain && color == ColorKeys.Warning}
    class:danger={!plain && color == ColorKeys.Danger}
    class:accent={!plain && color == ColorKeys.Accent}
    class:neutral={!plain && color == ColorKeys.Neutral}
    class:inherit={!plain && color == ColorKeys.Inherit}
    class:compact={compact}
    class:plain={plain}
    onclick={onClick}
    onmouseleave={(e) => {(e.target as HTMLButtonElement).blur()}}
    type={type}
    disabled={!enabled}
    class:disabled={!enabled}
    title={title || undefined}
    aria-label={title || undefined}
    onmousedown={addRipple}
    use:focusIndicator
  >
    {@render children?.()}
  </button>
{/if}
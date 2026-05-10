<script lang="ts" generics="T">

  import { NoOp } from "$lib/client/placeholders";
  import RadioToggle from "../interactive/RadioToggle.svelte";
  import type { Option } from "../../types/options";

  interface Props {
    value: T | null;
    name: string;
    editable?: boolean;
    options: Option<T>[];
    onClick?: (selected: T) => any;
  }

  let {
    value = $bindable(),
    name,
    editable = true,
    options,
    onClick = NoOp,
  }: Props = $props();
</script>

<style lang="scss">
  @use "../../styles/colors.scss";
  @use "../../styles/dimensions.scss";

  div {
    display: flex;
    align-items: start;
    flex-direction: row;
    flex-wrap: nowrap;
    gap: dimensions.$gapSmall;
    align-items: center;
    justify-content: start;
    cursor: pointer;
    width: max-content;
  }
  
  label {
    cursor: pointer;
    width: max-content;
    user-select: none;
    color: color-mix(in srgb, colors.$foregroundPrimary 50%, transparent);
  }
</style>

<!-- svelte-ignore a11y_click_events_have_key_events -->
{#each options as option (option.value)}
  <div
    role="radio"
    tabindex="-1"
    aria-checked={value == option.value}
  >
    <RadioToggle
      name={name}
      value={option.value}
      bind:selected={value}
      enabled={editable}
      onChange={(x) => { if (x !== null) onClick(x); }}
    />
    <label for={`${name}-${option.value}`}>
      {option.name}
    </label>
  </div>
{/each}
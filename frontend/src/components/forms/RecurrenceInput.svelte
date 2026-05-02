<script lang="ts">
  import { RRule, Weekday, type ByWeekday, type Options } from "rrule";
  import SelectInput from "./SelectInput.svelte";
  import ToggleInput from "./ToggleInput.svelte";
  import DateTimeInput from "./DateTimeInput.svelte";
  import { untrack } from "svelte";
  import SelectButtons from "./SelectButtons.svelte";
  import SelectButtonsMulti from "./SelectButtonsMulti.svelte";
  import { t } from "@sveltia/i18n";
  import { getDayName, getMonthName } from "$lib/common/humanization";
  import { getSettings } from "$lib/client/data/settings.svelte";
  import { UserSettingKeys } from "../../types/settings";
  import SelectInputMulti from "./SelectInputMulti.svelte";

  interface Props {
    options: Options;
    editable: boolean;
  }

  let {
    options = $bindable(),
    editable,
  }: Props = $props();

  const settings = getSettings();

  let byValsEnabled = $state({
    byMonthLimit: false,
    byYearDayLimit: false,
    byMonthDayLimit: false,
    byDayLimit: false,
    byHourLimit: false,
    byMinuteLimit: false,
    bySecondLimit: false,
    bySetPosLimit: false,
    byMonthExpand: false,
    byWeekNoExpand: false,
    byYearDayExpand: false,
    byMonthDayExpand: false,
    byDayExpand: false,
    byHourExpand: false,
    byMinuteExpand: false,
    bySecondExpand: false,
  })

  let byValsSemantics = $derived({
    byMonthLimit: options.freq != RRule.YEARLY,
    byYearDayLimit: [RRule.SECONDLY, RRule.MINUTELY, RRule.HOURLY].includes(options.freq),
    byMonthDayLimit: [RRule.SECONDLY, RRule.MINUTELY, RRule.HOURLY, RRule.DAILY].includes(options.freq),
    byDayLimit: [RRule.SECONDLY, RRule.MINUTELY, RRule.HOURLY, RRule.DAILY].includes(options.freq) || (options.freq == RRule.MONTHLY && byValsEnabled.byMonthDayExpand) || (options.freq == RRule.YEARLY && (byValsEnabled.byYearDayExpand || byValsEnabled.byMonthDayExpand)),
    byHourLimit: [RRule.SECONDLY, RRule.MINUTELY, RRule.HOURLY].includes(options.freq),
    byMinuteLimit: [RRule.SECONDLY, RRule.MINUTELY].includes(options.freq),
    bySecondLimit: [RRule.SECONDLY].includes(options.freq),
    byMonthExpand: options.freq == RRule.YEARLY,
    byWeekNoExpand: options.freq == RRule.YEARLY,
    byYearDayExpand: options.freq == RRule.YEARLY,
    byMonthDayExpand: [RRule.MONTHLY, RRule.YEARLY].includes(options.freq),
    byDayExpand: options.freq == RRule.WEEKLY || (options.freq == RRule.MONTHLY && !byValsEnabled.byMonthDayExpand) || (options.freq == RRule.YEARLY && !byValsEnabled.byYearDayExpand && !byValsEnabled.byMonthDayExpand),
    byHourExpand: [RRule.DAILY, RRule.WEEKLY, RRule.MONTHLY, RRule.YEARLY].includes(options.freq),
    byMinuteExpand: [RRule.HOURLY, RRule.DAILY, RRule.WEEKLY, RRule.MONTHLY, RRule.YEARLY].includes(options.freq),
    bySecondExpand: options.freq != RRule.SECONDLY,
    byDayNumeric: options.freq == RRule.MONTHLY || (options.freq == RRule.YEARLY && !byValsEnabled.byWeekNoExpand)
  })

  let byVals = $state({
    byMonthLimit: [] as number[],
    byYearDayLimit: [] as number[],
    byMonthDayLimit: [] as number[],
    byDayLimit: [] as ByWeekday[],
    byHourLimit: [] as number[],
    byMinuteLimit: [] as number[],
    bySecondLimit: [] as number[],
    bySetPosLimit: [] as number[],
    byMonthExpand: [] as number[],
    byWeekNoExpand: [] as number[],
    byYearDayExpand: [] as number[],
    byMonthDayExpand: [] as number[],
    byDayExpand: [] as ByWeekday[],
    byHourExpand: [] as number[],
    byMinuteExpand: [] as number[],
    bySecondExpand: [] as number[],
  })

  // Initialize values
  $effect(() => {
    options;
    untrack(() => {
      // Initialize BYMONTH
      if (options.bymonth == null) {
        byValsEnabled.byMonthExpand = false;
        byValsEnabled.byMonthLimit = false;
      } else {
        const arr = Array.isArray(options.bymonth) ? options.bymonth : [ options.bymonth ];
        if (byValsSemantics.byMonthExpand) {
          byValsEnabled.byMonthExpand = true;
          byVals.byMonthExpand = arr;
        } else if (byValsSemantics.byMonthLimit) {
          byValsEnabled.byMonthLimit = true;
          byVals.byMonthLimit = arr;
        }
      }

      // Initialize BYWEEKNO
      if (options.byweekno == null) {
        byValsEnabled.byWeekNoExpand = false;
      } else {
        const arr = Array.isArray(options.byweekno) ? options.byweekno : [ options.byweekno ];
        if (byValsSemantics.byWeekNoExpand) {
          byValsEnabled.byWeekNoExpand = true;
          byVals.byWeekNoExpand = arr;
        }
      }

      // Initialize BYYEARDAY
      if (options.byyearday == null) {
        byValsEnabled.byYearDayExpand = false;
        byValsEnabled.byYearDayLimit = false;
      } else {
        const arr = Array.isArray(options.byyearday) ? options.byyearday : [ options.byyearday ];
        if (byValsSemantics.byYearDayExpand) {
          byValsEnabled.byYearDayExpand = true;
          byVals.byYearDayExpand = arr;
        } else if (byValsSemantics.byYearDayLimit) {
          byValsEnabled.byYearDayLimit = true;
          byVals.byYearDayLimit = arr;
        }
      }

      // Initialize BYMONTHDAY
      if (options.bymonthday == null) {
        byValsEnabled.byMonthDayExpand = false;
        byValsEnabled.byMonthDayLimit = false;
      } else {
        const arr = Array.isArray(options.bymonthday) ? options.bymonthday : [ options.bymonthday ];
        if (byValsSemantics.byMonthDayExpand) {
          byValsEnabled.byMonthDayExpand = true;
          byVals.byMonthDayExpand = arr;
        } else if (byValsSemantics.byMonthDayLimit) {
          byValsEnabled.byMonthDayLimit = true;
          byVals.byMonthDayLimit = arr;
        }
      }

      // Initialize BYDAY
      if (options.byweekday == null) {
        byValsEnabled.byDayExpand = false;
        byValsEnabled.byDayLimit = false;
      } else {
        const arr = Array.isArray(options.byweekday) ? options.byweekday : [ options.byweekday ];
        if (byValsSemantics.byDayExpand) {
          byValsEnabled.byDayExpand = true;
          byVals.byDayExpand = arr;
        } else if (byValsSemantics.byDayLimit) {
          byValsEnabled.byDayLimit = true;
          byVals.byDayLimit = arr;
        }
      }

      // Initialize BYHOUR
      if (options.byhour == null) {
        byValsEnabled.byHourExpand = false;
        byValsEnabled.byHourLimit = false;
      } else {
        const arr = Array.isArray(options.byhour) ? options.byhour : [ options.byhour ];
        if (byValsSemantics.byHourExpand) {
          byValsEnabled.byHourExpand = true;
          byVals.byHourExpand = arr;
        } else if (byValsSemantics.byHourLimit) {
          byValsEnabled.byHourLimit = true;
          byVals.byHourLimit = arr;
        }
      }

      // Initialize BYMINUTE
      if (options.byminute == null) {
        byValsEnabled.byMinuteExpand = false;
        byValsEnabled.byMinuteLimit = false;
      } else {
        const arr = Array.isArray(options.byminute) ? options.byminute : [ options.byminute ];
        if (byValsSemantics.byMinuteExpand) {
          byValsEnabled.byMinuteExpand = true;
          byVals.byMinuteExpand = arr;
        } else if (byValsSemantics.byMinuteLimit) {
          byValsEnabled.byMinuteLimit = true;
          byVals.byMinuteLimit = arr;
        }
      }

      // Initialize BYSECOND
      if (options.bysecond == null) {
        byValsEnabled.bySecondExpand = false;
        byValsEnabled.bySecondLimit = false;
      } else {
        const arr = Array.isArray(options.bysecond) ? options.bysecond : [ options.bysecond ];
        if (byValsSemantics.bySecondExpand) {
          byValsEnabled.bySecondExpand = true;
          byVals.bySecondExpand = arr;
        } else if (byValsSemantics.bySecondLimit) {
          byValsEnabled.bySecondLimit = true;
          byVals.bySecondLimit = arr;
        }
      }

      // Initialize BYSETPOS
      if (options.bysetpos == null) {
        byValsEnabled.bySetPosLimit = false;
      } else {
        const arr = Array.isArray(options.bysetpos) ? options.bysetpos : [ options.bysetpos ];
        if (byValsSemantics.bySecondLimit) {
          byValsEnabled.bySetPosLimit = true;
          byVals.bySetPosLimit = arr;
        }
      }
    })
  })

  let yearlyType: "date" | "day" = $state("date");
  let endType: "date" | "count" | "forever" = $state("forever");

  $effect(() => {
    endType;
    untrack(() => {
      switch (endType) {
        case "date":
          options.count = null;
          if (options.until == null) options.until = new Date();
          break
        case "count":
          options.until = null;
          if (options.count == null) options.count = 1;
          break
        case "forever":
          options.count = null;
          options.until = null;
          break
      }
    })
  });

  // Set WKST
  $effect(() => {
    const wkst = settings.userSettings[UserSettingKeys.FirstDayOfWeek];
    untrack(() => {
      options.wkst = [RRule.SU, RRule.MO, RRule.TU, RRule.WE, RRule.TH, RRule.FR, RRule.SA][wkst];
    })
  })

  // Set BYMONTH
  $effect(() => {
    byValsSemantics.byMonthExpand,
    byValsSemantics.byMonthLimit,
    byValsEnabled.byMonthExpand,
    byValsEnabled.byMonthLimit,
    byVals.byMonthExpand,
    byVals.byMonthLimit
    untrack(() => {
      if (byValsSemantics.byMonthExpand && byValsEnabled.byMonthExpand) options.bymonth = byVals.byMonthExpand;
      else if (byValsSemantics.byMonthLimit && byValsEnabled.byMonthLimit) options.bymonth = byVals.byMonthLimit;
      else options.bymonth = null;
    })
  })

  // Set BYWEEKNO
  $effect(() => {
    byValsSemantics.byWeekNoExpand,
    byValsEnabled.byWeekNoExpand,
    byVals.byWeekNoExpand,
    untrack(() => {
      if (byValsSemantics.byWeekNoExpand && byValsEnabled.byWeekNoExpand) options.byweekno = byVals.byWeekNoExpand;
      else options.byweekno = null;
    })
  })

  // Set BYYEARDAY
  $effect(() => {
    byValsSemantics.byYearDayExpand,
    byValsSemantics.byYearDayLimit,
    byValsEnabled.byYearDayExpand,
    byValsEnabled.byYearDayLimit,
    byVals.byYearDayExpand,
    byValsSemantics.byYearDayLimit
    untrack(() => {
      if (byValsSemantics.byYearDayExpand && byValsEnabled.byYearDayExpand) options.byyearday = byVals.byYearDayExpand;
      else if (byValsSemantics.byYearDayLimit && byValsEnabled.byYearDayLimit) options.byyearday = byVals.byYearDayLimit;
      else options.byyearday = null;
    })
  })

  // Set BYMONTHDAY
  $effect(() => {
    byValsSemantics.byMonthDayExpand,
    byValsSemantics.byMonthDayLimit,
    byValsEnabled.byMonthDayExpand,
    byValsEnabled.byMonthDayLimit,
    byVals.byMonthDayExpand,
    byValsSemantics.byMonthDayLimit
    untrack(() => {
      options.bynmonthday = null;
      if (byValsSemantics.byMonthDayExpand && byValsEnabled.byMonthDayLimit) options.bymonthday = byVals.byMonthDayExpand;
      else if (byValsSemantics.byMonthDayExpand && byValsEnabled.byMonthDayLimit) options.bymonthday = byVals.byMonthDayLimit;
      else options.bymonthday = null;
    })
  })

  // Set BYDAY
  $effect(() => {
    byValsSemantics.byDayExpand,
    byValsSemantics.byDayLimit,
    byValsEnabled.byDayExpand,
    byValsEnabled.byDayLimit,
    byVals.byDayExpand,
    byValsSemantics.byDayLimit
    untrack(() => {
      options.bynweekday = null;
      if (byValsSemantics.byDayExpand && byValsEnabled.byDayExpand) options.byweekday = byVals.byDayExpand;
      else if (byValsSemantics.byDayLimit && byValsEnabled.byDayLimit) options.byweekday = byVals.byDayLimit;
      else options.byweekday = null;
    })
  })

  // Set BYHOUR
  $effect(() => {
    byValsSemantics.byHourExpand,
    byValsSemantics.byHourLimit,
    byValsEnabled.byHourExpand,
    byValsEnabled.byHourLimit,
    byVals.byHourExpand,
    byValsSemantics.byHourLimit
    untrack(() => {
      if (byValsSemantics.byHourExpand && byValsEnabled.byHourExpand) options.byhour = byVals.byHourExpand;
      else if (byValsSemantics.byHourLimit && byValsEnabled.byHourLimit) options.byhour = byVals.byHourLimit;
      else options.byhour = null;
    })
  })

  // Set BYMINUTE
  $effect(() => {
    byValsSemantics.byMinuteExpand,
    byValsSemantics.byMinuteLimit,
    byValsEnabled.byMinuteExpand,
    byValsEnabled.byMinuteLimit,
    byVals.byMinuteExpand,
    byValsSemantics.byMinuteLimit
    untrack(() => {
      if (byValsSemantics.byMinuteExpand && byValsEnabled.byMinuteExpand) options.byminute = byVals.byMinuteExpand;
      else if (byValsSemantics.byMinuteLimit && byValsEnabled.byMinuteLimit) options.byminute = byVals.byMinuteLimit;
      else options.byminute = null;
    })
  })
  
  // Set BYSECOND
  $effect(() => {
    byValsSemantics.bySecondExpand,
    byValsSemantics.bySecondLimit,
    byValsEnabled.bySecondExpand,
    byValsEnabled.bySecondLimit,
    byVals.bySecondExpand,
    byValsSemantics.bySecondLimit
    untrack(() => {
      if (byValsSemantics.bySecondExpand && byValsEnabled.bySecondExpand) options.bysecond = byVals.bySecondExpand;
      else if (byValsSemantics.bySecondLimit && byValsEnabled.bySecondLimit) options.bysecond = byVals.bySecondLimit;
      else options.bysecond = null;
    })
  })
</script>

<!-- FREQ -->
<SelectInput bind:value={options.freq} name="recurrence_freq" placeholder="Frequency" showLabel={true} options={[
  { value: RRule.SECONDLY, name: "Secondly" },
  { value: RRule.MINUTELY, name: "Minutely" },
  { value: RRule.HOURLY, name: "Hourly" },
  { value: RRule.DAILY, name: "Daily" },
  { value: RRule.WEEKLY, name: "Weekly" },
  { value: RRule.MONTHLY, name: "Monthly" },
  { value: RRule.YEARLY, name: "Yearly" },
]} editable={editable} />

<!--
{#if options.freq == RRule.YEARLY}
  <SelectButtons bind:value={yearlyType} name="recurrence_freq_yearly_type" placeholder="By" options={[
    { value: "date", name: "By day number" },
    { value: "day", name: "By week day" },
  ]} editable={editable} />
{/if}
-->

<!-- BYMONTH -->
{#if byValsSemantics.byMonthLimit }
  <ToggleInput
    bind:value={byValsEnabled.byMonthLimit}
    name="recurrence_bymonth_limit_enable" 
    description="Only during specific months"
  />
  {#if byValsEnabled.byMonthLimit}
    <SelectButtonsMulti
      bind:values={byVals.byMonthLimit}
      name="recurrence_bymonth_limit"
      placeholder="Months"
      options={ [...Array(12).keys()].map(x => ({ value: x+1, name: getMonthName(x, true) })) }
      onClick={() => { byVals.byMonthLimit = byVals.byMonthLimit; }}
    />
  {/if}
{/if}
{#if byValsSemantics.byMonthExpand }
  <ToggleInput
    bind:value={byValsEnabled.byMonthExpand}
    name="recurrence_bymonth_expand_enable" 
    description="Multiple months per year"
  />
  {#if byValsEnabled.byMonthExpand}
    <SelectButtonsMulti
      bind:values={byVals.byMonthExpand}
      name="recurrence_bymonth_expand"
      placeholder="Months"
      options={ [...Array(12).keys()].map(x => ({ value: x+1, name: getMonthName(x, true) })) }
      onClick={() => { byVals.byMonthExpand = byVals.byMonthExpand; }}
    />
  {/if}
{/if}

<!-- BYWEEKNO -->
{#if byValsSemantics.byWeekNoExpand }
  <ToggleInput
    bind:value={byValsEnabled.byWeekNoExpand}
    name="recurrence_byweekno_expand_enable" 
    description="Multiple weeks per year"
  />
  {#if byValsEnabled.byWeekNoExpand}
    <SelectInputMulti
      bind:values={byVals.byWeekNoExpand}
      name="recurrence_byweekno_expand"
      placeholder="Week numbers"
      options={
        [...Array(53).keys()].map(x => ({ value: x+1, name: t("numbers.ordinal.normal", { values: { num: x + 1 } }) })).concat(
        [...Array(53).keys()].map(x => ({ value: -(x+1), name: t("numbers.ordinal.reverse", { values: { num: x + 1 } }) })))
      }
      click={() => { byVals.byWeekNoExpand = byVals.byWeekNoExpand; }}
    />
  {/if}
{/if}

<!-- BYYEARDAY -->
{#if byValsSemantics.byYearDayLimit }
  <ToggleInput
    bind:value={byValsEnabled.byYearDayLimit}
    name="recurrence_byyearday_limit_enable" 
    description="Only on specific days of the year"
  />
  {#if byValsEnabled.byYearDayLimit}
    <SelectInputMulti
      bind:values={byVals.byMonthDayLimit}
      name="recurrence_byyearday_limit"
      placeholder="Days of the year"
      options={
        [...Array(366).keys()].map(x => ({ value: x+1, name: t("numbers.ordinal.normal", { values: { num: x + 1 } }) })).concat(
        [...Array(366).keys()].map(x => ({ value: -(x+1), name: t("numbers.ordinal.reverse", { values: { num: x + 1 } }) })))
      }
      click={() => { byVals.byYearDayLimit = byVals.byYearDayLimit; }}
    />
  {/if}
{/if}
{#if byValsSemantics.byYearDayExpand }
  <ToggleInput
    bind:value={byValsEnabled.byYearDayExpand}
    name="recurrence_byyearday_limit_enable" 
    description="Multiple days per year"
  />
  {#if byValsEnabled.byYearDayExpand}
    <SelectInputMulti
      bind:values={byVals.byMonthDayExpand}
      name="recurrence_byyearday_expand"
      placeholder="Days of the year"
      options={
        [...Array(366).keys()].map(x => ({ value: x+1, name: t("numbers.ordinal.normal", { values: { num: x + 1 } }) })).concat(
        [...Array(366).keys()].map(x => ({ value: -(x+1), name: t("numbers.ordinal.reverse", { values: { num: x + 1 } }) })))
      }
      click={() => { byVals.byYearDayExpand = byVals.byYearDayExpand; }}
    />
  {/if}
{/if}

<!-- BYMONTHDAY -->
{#if byValsSemantics.byMonthDayLimit }
  <ToggleInput
    bind:value={byValsEnabled.byMonthDayLimit}
    name="recurrence_bymonth_limit_enable" 
    description="Only on specific days of the month"
  />
  {#if byValsEnabled.byMonthDayLimit}
    <SelectInputMulti
      bind:values={byVals.byMonthDayLimit}
      name="recurrence_bymonthday_limit"
      placeholder="Days of the month"
      options={
        [...Array(31).keys()].map(x => ({ value: x+1, name: t("numbers.ordinal.normal", { values: { num: x + 1 } }) })).concat(
        [...Array(31).keys()].map(x => ({ value: -(x+1), name: t("numbers.ordinal.reverse", { values: { num: x + 1 } }) })))
      }
      click={() => { byVals.byMonthDayLimit = byVals.byMonthDayLimit; }}
    />
  {/if}
{/if}
{#if byValsSemantics.byMonthDayExpand }
  <ToggleInput
    bind:value={byValsEnabled.byMonthDayExpand}
    name="recurrence_bymonth_expand_enable" 
    description="Multiple days per month"
  />
  {#if byValsEnabled.byMonthDayExpand}
    <SelectInputMulti
      bind:values={byVals.byMonthDayExpand}
      name="recurrence_bymonthday_expand"
      placeholder="Days of the month"
      options={
        [...Array(31).keys()].map(x => ({ value: x+1, name: t("numbers.ordinal.normal", { values: { num: x + 1 } }) })).concat(
        [...Array(31).keys()].map(x => ({ value: -(x+1), name: t("numbers.ordinal.reverse", { values: { num: x + 1 } }) })))
      }
      click={() => { byVals.byMonthDayExpand = byVals.byMonthDayExpand; }}
    />
  {/if}
{/if}

<!-- BYDAY -->
{#if byValsSemantics.byDayLimit }
  <ToggleInput
    bind:value={byValsEnabled.byDayLimit}
    name="recurrence_byday_limit_enable" 
    description="Only on specific weekdays"
  />
  {#if byValsEnabled.byDayLimit}
    {#if byValsSemantics.byDayNumeric}
      <SelectInputMulti
        bind:values={byVals.byDayLimit}
        name="recurrence_byday_limit"
        placeholder="Weekdays"
        options={
          [RRule.SU, RRule.MO, RRule.TU, RRule.WE, RRule.TH, RRule.FR, RRule.SA]
            .map((x, i) => ({ value: x, name: getDayName(i, false), index: (i + 7 - settings.userSettings[UserSettingKeys.FirstDayOfWeek]) % 7 }))
            .sort((a, b) =>  a.index-b.index)
            .flatMap(x =>
              [{ value: x.value, name: `Every ${x.name}`}].concat(
                [...Array(4).keys()].map(y => ({ value: x.value.nth(y+1), name: `${t("numbers.ordinal.normal", { values: { num: y + 1 } })} ${x.name}` }))
              ).concat(
                [...Array(4).keys()].map(y => ({ value: x.value.nth(-(y+1)), name: `${t("numbers.ordinal.reverse", { values: { num: y + 1 } })} ${x.name}` }))
              )
            )
        }
        click={() => { byVals.byDayLimit = byVals.byDayLimit; }}
      />
    {:else}
      <SelectButtonsMulti
        bind:values={byVals.byDayLimit}
        name="recurrence_byday_limit"
        placeholder="Weekdays"
        options={
          [RRule.SU, RRule.MO, RRule.TU, RRule.WE, RRule.TH, RRule.FR, RRule.SA]
            .map((x, i) => ({ value: x, name: getDayName(i, true), index: (i + 7 - settings.userSettings[UserSettingKeys.FirstDayOfWeek]) % 7 }))
            .sort((a, b) =>  a.index-b.index)
        }
        onClick={() => { byVals.byDayLimit = byVals.byDayLimit; }}
      />
    {/if}
  {/if}
{/if}
{#if byValsSemantics.byDayExpand }
  <ToggleInput
    bind:value={byValsEnabled.byDayExpand}
    name="recurrence_byday_expand_enable" 
    description={byValsSemantics.byDayNumeric ? "Multiple weekdays" : "Multiple days per week"}
  />
  {#if byValsEnabled.byDayExpand}
    {#if byValsSemantics.byDayNumeric}
      <SelectInputMulti
        bind:values={byVals.byDayExpand}
        name="recurrence_byday_expand"
        placeholder="Weekdays"
        options={
          [RRule.SU, RRule.MO, RRule.TU, RRule.WE, RRule.TH, RRule.FR, RRule.SA]
            .map((x, i) => ({ value: x, name: getDayName(i, false), index: (i + 7 - settings.userSettings[UserSettingKeys.FirstDayOfWeek]) % 7 }))
            .sort((a, b) =>  a.index-b.index)
            .flatMap(x =>
              [{ value: x.value, name: `Every ${x.name}`}].concat(
                [...Array(4).keys()].map(y => ({ value: x.value.nth(y+1), name: `${t("numbers.ordinal.normal", { values: { num: y + 1 } })} ${x.name}` }))
              ).concat(
                [...Array(4).keys()].map(y => ({ value: x.value.nth(-(y+1)), name: `${t("numbers.ordinal.reverse", { values: { num: y + 1 } })} ${x.name}` }))
              )
            )
        }
        click={() => { byVals.byDayExpand = byVals.byDayExpand; }}
      />
    {:else}
      <SelectButtonsMulti
        bind:values={byVals.byDayExpand}
        name="recurrence_byday_expand"
        placeholder="Weekdays"
        options={
          [RRule.SU, RRule.MO, RRule.TU, RRule.WE, RRule.TH, RRule.FR, RRule.SA]
            .map((x, i) => ({ value: x, name: getDayName(i, true), index: (i + 7 - settings.userSettings[UserSettingKeys.FirstDayOfWeek]) % 7 }))
            .sort((a, b) =>  a.index-b.index)
        }
        onClick={() => { byVals.byDayExpand = byVals.byDayExpand; }}
      />
    {/if}
  {/if}
{/if}

<!-- BYHOUR -->
{#if byValsSemantics.byHourLimit }
  <ToggleInput
    bind:value={byValsEnabled.byHourLimit}
    name="recurrence_byhour_limit_enable" 
    description="Only on specific hours of the day"
  />
  {#if byValsEnabled.byHourLimit}
    <SelectInputMulti
      bind:values={byVals.byHourLimit}
      name="recurrence_byhour_limit"
      placeholder="Hours of the day"
      options={
        [...Array(24).keys()].map(x => ({ value: x+1, name: t("numbers.ordinal.normal", { values: { num: x + 1} }) })).concat(
        [...Array(24).keys()].map(x => ({ value: -(x+1), name: t("numbers.ordinal.reverse", { values: { num: x + 1} }) })))
      }
      click={() => { byVals.byHourLimit = byVals.byHourLimit; }}
    />
  {/if}
{/if}
{#if byValsSemantics.byHourExpand }
  <ToggleInput
    bind:value={byValsEnabled.byHourExpand}
    name="recurrence_byhour_expand_enable" 
    description="Multiple hours per day"
  />
  {#if byValsEnabled.byHourExpand}
    <SelectInputMulti
      bind:values={byVals.byHourExpand}
      name="recurrence_byhour_expand"
      placeholder="Hours of the day"
      options={
        [...Array(24).keys()].map(x => ({ value: x+1, name: t("numbers.ordinal.normal", { values: { num: x + 1} }) })).concat(
        [...Array(24).keys()].map(x => ({ value: -(x+1), name: t("numbers.ordinal.reverse", { values: { num: x + 1} }) })))
      }
      click={() => { byVals.byHourExpand = byVals.byHourExpand; }}
    />
  {/if}
{/if}

<!-- BYMINUTE -->
{#if byValsSemantics.byMinuteLimit }
  <ToggleInput
    bind:value={byValsEnabled.byMinuteLimit}
    name="recurrence_byminute_limit_enable" 
    description="Only on specific minutes of the hour"
  />
  {#if byValsEnabled.byMinuteLimit}
    <SelectInputMulti
      bind:values={byVals.byMinuteLimit}
      name="recurrence_byminute_limit"
      placeholder="Minutes of the hour"
      options={
        [...Array(60).keys()].map(x => ({ value: x+1, name: t("numbers.ordinal.normal", { values: { num: x + 1} }) })).concat(
        [...Array(60).keys()].map(x => ({ value: -(x+1), name: t("numbers.ordinal.reverse", { values: { num: x + 1} }) })))
      }
      click={() => { byVals.byMinuteLimit = byVals.byMinuteLimit; }}
    />
  {/if}
{/if}
{#if byValsSemantics.byMinuteExpand }
  <ToggleInput
    bind:value={byValsEnabled.byMinuteExpand}
    name="recurrence_byminute_expand_enable" 
    description="Multiple minutes per hour"
  />
  {#if byValsEnabled.byMinuteExpand}
    <SelectInputMulti
      bind:values={byVals.byMinuteExpand}
      name="recurrence_byminute_expand"
      placeholder="Minutes of the hour"
      options={
        [...Array(60).keys()].map(x => ({ value: x+1, name: t("numbers.ordinal.normal", { values: { num: x + 1} }) })).concat(
        [...Array(60).keys()].map(x => ({ value: -(x+1), name: t("numbers.ordinal.reverse", { values: { num: x + 1} }) })))
      }
      click={() => { byVals.byMinuteExpand = byVals.byMinuteExpand; }}
    />
  {/if}
{/if}

<!-- BYSECOND -->
{#if byValsSemantics.bySecondLimit }
  <ToggleInput
    bind:value={byValsEnabled.bySecondLimit}
    name="recurrence_bysecond_limit_enable" 
    description="Only on specific seconds of the minute"
  />
  {#if byValsEnabled.bySecondLimit}
    <SelectInputMulti
      bind:values={byVals.bySecondLimit}
      name="recurrence_bysecond_limit"
      placeholder="Seconds of the minute"
      options={
        [...Array(60).keys()].map(x => ({ value: x+1, name: t("numbers.ordinal.normal", { values: { num: x + 1} }) })).concat(
        [...Array(60).keys()].map(x => ({ value: -(x+1), name: t("numbers.ordinal.reverse", { values: { num: x + 1} }) })))
      }
      click={() => { byVals.bySecondLimit = byVals.bySecondLimit; }}
    />
  {/if}
{/if}
{#if byValsSemantics.bySecondExpand }
  <ToggleInput
    bind:value={byValsEnabled.bySecondExpand}
    name="recurrence_bysecond_expand_enable" 
    description="Multiple seconds per minute"
  />
  {#if byValsEnabled.bySecondExpand}
    <SelectInputMulti
      bind:values={byVals.bySecondExpand}
      name="recurrence_bysecond_expand"
      placeholder="Seconds of the minute"
      options={
        [...Array(60).keys()].map(x => ({ value: x+1, name: t("numbers.ordinal.normal", { values: { num: x + 1} }) })).concat(
        [...Array(60).keys()].map(x => ({ value: -(x+1), name: t("numbers.ordinal.reverse", { values: { num: x + 1} }) })))
      }
      click={() => { byVals.bySecondExpand = byVals.bySecondExpand; }}      
    />
  {/if}
{/if}

<SelectButtons bind:value={endType} name="recurrence_duration" placeholder="Duration" options={[
  { value: "date", name: "Until specified date" },
  { value: "count", name: "Repetion count" },
  { value: "forever", name: "Repeat forever" },
]} editable={editable} />

{#if endType == "date"}
  <DateTimeInput
    bind:value={options.until}
    editable={editable}
    placeholder="Until"
    name="recurrence_until" 
  />
{/if}

{(new RRule(options)).toString()}
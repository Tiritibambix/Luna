<script lang="ts">
  import Loader from "../decoration/Loader.svelte";
  import Button from "../interactive/Button.svelte";
  import Horizontal from "../layout/Horizontal.svelte";
  import Modal from "./Modal.svelte";
  import MonthSelection from "../interactive/MonthSelection.svelte";
  import SmallCalendar from "../interactive/SmallCalendar.svelte";

  import { isSameDay } from "$lib/common/date";
  import { NoOp } from "$lib/client/placeholders";
  import { queueNotification } from "$lib/client/notifications";
  import { ColorKeys } from "../../types/colors";

  interface Props {
    onConfirm: (dates: Date[]) => Promise<void>;
    showModal?: () => any;
    hideModal?: () => any;
  }

  let {
    showModal = $bindable(),
    hideModal = $bindable(NoOp),
    onConfirm,
  }: Props = $props();

  let showModalInternal: () => any = $state(NoOp);
  let hideModalInternal: () => any = $state(NoOp);

  let viewDate: Date = $state(new Date());
  let selectedDates: Date[] = $state([]);

  showModal = () => {
    viewDate = new Date();
    selectedDates = [];
    showModalInternal();
  };

  function toggleDate(day: Date) {
    const index = selectedDates.findIndex(d => isSameDay(d, day));
    if (index === -1) selectedDates = [...selectedDates, day];
    else selectedDates = selectedDates.filter((_, i) => i !== index);
  }

  function isSelected(day: Date): boolean {
    return selectedDates.some(d => isSameDay(d, day));
  }

  let awaitingConfirm = $state(false);
  function confirm() {
    if (selectedDates.length === 0) return;
    awaitingConfirm = true;
    onConfirm(selectedDates).catch(err => {
      queueNotification(ColorKeys.Danger, err);
    }).finally(() => {
      awaitingConfirm = false;
      hideModalInternal();
    });
  }
</script>

<Modal title="Copy to dates" bind:showModal={showModalInternal} bind:hideModal={hideModalInternal}>
  <MonthSelection bind:date={viewDate} />
  <SmallCalendar bind:date={viewDate} onDayClick={toggleDate} isSelected={isSelected} dynamicRows={false} />
  <Horizontal position="left">
    {selectedDates.length} date{selectedDates.length === 1 ? "" : "s"} selected
  </Horizontal>
  {#snippet buttons()}
    <Button onClick={confirm} color={ColorKeys.Success} enabled={selectedDates.length > 0} type="submit">
      {#if awaitingConfirm}
        <Loader/>
      {:else}
        Confirm
      {/if}
    </Button>
    <Button onClick={hideModalInternal} color={ColorKeys.Danger}>Cancel</Button>
  {/snippet}
</Modal>

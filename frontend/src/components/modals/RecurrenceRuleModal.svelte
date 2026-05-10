<script lang="ts">
  import Modal from "./Modal.svelte";

  import { NoOp } from "../../lib/client/placeholders";
  import { RRule, type Options } from "rrule";
  import RecurrenceInput from "../forms/RecurrenceInput.svelte";

  interface Props {
    showModal: (initial: Options) => Promise<Options>;
    dtstart: Date;
    allDay: boolean; }

  let success: (result: Options) => void = $state(NoOp);
  let failure: (reason?: string | Error) => void = $state(NoOp);

  let {
    showModal = $bindable(),
    dtstart,
    allDay
  }: Props = $props();

  let showModalInternal: () => Promise<Options> = $state(Promise.reject);
  let options = $state<Options>((new RRule()).options)

  showModal = async (initial) => {
    options = initial;
    return showModalInternal();
  }
</script>

<Modal title={"Recurrence editing"} bind:showModal={showModalInternal} bind:success bind:failure>
  <RecurrenceInput
    dtstart={dtstart} 
    options={options}
    allDay={allDay}
    editable={true}
  />
</Modal>
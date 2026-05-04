<script lang="ts">
  import { t } from "@sveltia/i18n";
  import type { Settings } from "../../../lib/client/data/settings.svelte";
  import { ColorKeys } from "../../../types/colors";
  import Button from "../../interactive/Button.svelte";
  import { downloadFileToClient, fetchFile } from "$lib/client/net";
  import { queueNotification } from "$lib/client/notifications";

  interface Props {
    settings: Settings;
    showConfirmation: (message: string, details?: string) => Promise<void>;
  }

  let {
    settings,
    showConfirmation,
  }: Props = $props();

  async function createBackup() {
    await fetchFile("/api/backups/create", { method: "POST" }).then(downloadFileToClient).catch(e => {
      queueNotification(ColorKeys.Danger, e);
    })
  }
</script>

<Button color={ColorKeys.Accent} onClick={createBackup}>{t("settings.backups.create")}</Button>
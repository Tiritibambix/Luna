<script lang="ts">
  import { t } from "@sveltia/i18n";
  import type { Settings } from "../../../lib/client/data/settings.svelte";
  import { ColorKeys } from "../../../types/colors";
  import Button from "../../interactive/Button.svelte";
  import { downloadFileToClient, fetchFile, fetchResponse } from "$lib/client/net";
  import { queueNotification } from "$lib/client/notifications";
  import FileUpload from "../../forms/FileUpload.svelte";

  interface Props {
    settings: Settings;
    showConfirmation: (message: string, details?: string, notice?: boolean) => Promise<void>;
  }

  let {
    settings,
    showConfirmation,
  }: Props = $props();

  let backupFile: FileList | null = $state(null);

  async function createBackup() {
    await fetchFile("/api/backups/create", { method: "POST" }).then(downloadFileToClient).catch(e => {
      queueNotification(ColorKeys.Danger, e);
    })
  }

  async function restoreBackup() {
    console.log(backupFile);
    if (backupFile === null) {
      queueNotification(ColorKeys.Danger, t("settings.backups.error.file"));
      return;
    }

    const confirmed = await showConfirmation(`${t("settings.backups.confirm")}\n${t("confirmation.irreversible")}`).then(() => true).catch(() => false);
    if (!confirmed) return;

    const formData = new FormData();
    formData.append("backup_file", backupFile[0]);

    return fetchResponse("/api/backups/restore", { method: "POST", body: formData }).then(() =>
      showConfirmation(t("settings.backups.restart"), "", true)
    ).catch(err => {
      queueNotification(ColorKeys.Danger, t("settings.backups.error.restore", { values: { msg: err.message } }));
    })
  }
</script>

<Button color={ColorKeys.Accent} onClick={createBackup}>{t("settings.backups.create")}</Button>

<FileUpload
  name="backup_file"
  placeholder={t("settings.backups.file")}
  bind:files={backupFile}
  accept={".tar.gz"}
/>
{#if backupFile !== null}
  <Button color={ColorKeys.Success} onClick={restoreBackup}>
    {t("settings.backups.restore")}
  </Button>
{/if}
import { toast } from "vue-sonner";
import type { UploadedFile } from "./useFileAttach";

/**
 * Holds attachments picked while a task is still being composed. Files are
 * uploaded immediately (uploads are task-independent) and linked to the task
 * once it exists.
 */
export function usePendingAttachments() {
  const { uploadFiles, uploading } = useFileAttach();
  const { attachFile } = useAttachments();

  const pending = ref<UploadedFile[]>([]);

  async function addFiles(files: File[]) {
    if (!files.length) return;
    const uploaded = await uploadFiles(files);
    pending.value.push(...uploaded);

    const failed = files.length - uploaded.length;
    if (failed > 0) {
      toast.error(`${failed} file${failed > 1 ? "s" : ""} failed to upload`);
    }
  }

  function remove(uploadId: string) {
    pending.value = pending.value.filter((f) => f.uploadId !== uploadId);
  }

  function clear() {
    pending.value = [];
  }

  async function attachAll(projectKey: string, taskNum: number) {
    for (const file of pending.value) {
      const result = await attachFile(projectKey, taskNum, "task", file.uploadId);
      if (!result.success) {
        toast.error(`Failed to attach ${file.filename}`);
      }
    }
    clear();
  }

  return { pending, uploading, addFiles, remove, clear, attachAll };
}

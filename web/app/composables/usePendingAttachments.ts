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
    // Per-file upload failures are reported by useFileAttach().
    pending.value.push(...(await uploadFiles(files)));
  }

  function remove(uploadId: string) {
    pending.value = pending.value.filter((f) => f.uploadId !== uploadId);
  }

  function clear() {
    pending.value = [];
  }

  // Passing a commentId links the files to that comment instead of the task.
  async function attachAll(projectKey: string, taskNum: number, commentId?: string) {
    for (const file of pending.value) {
      const result = await attachFile(
        projectKey,
        taskNum,
        commentId ? "comment" : "task",
        file.uploadId,
        commentId
      );
      if (!result.success) {
        toast.error(`Failed to attach ${file.filename}`);
      }
    }
    clear();
  }

  return { pending, uploading, addFiles, remove, clear, attachAll };
}

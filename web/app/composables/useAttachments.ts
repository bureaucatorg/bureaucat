export interface Attachment {
  id: string;
  upload_id: string;
  filename: string;
  mime_type: string;
  size_bytes: number;
  url: string;
  created_by: string;
  created_at: string;
}

export function useAttachments() {
  const { getAuthHeader } = useAuth();

  async function listAttachments(
    projectKey: string,
    taskNum: number,
    entityType: "task" | "comment",
    commentId?: string
  ): Promise<{ success: boolean; data?: Attachment[]; error?: string }> {
    try {
      const base = `/api/v1/projects/${projectKey}/tasks/${taskNum}`;
      const url =
        entityType === "task"
          ? `${base}/attachments`
          : `${base}/comments/${commentId}/attachments`;

      const response = await fetch(url, {
        headers: getAuthHeader(),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message };
      }

      const data: Attachment[] = await response.json();
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function attachFile(
    projectKey: string,
    taskNum: number,
    entityType: "task" | "comment",
    uploadId: string,
    commentId?: string
  ): Promise<{ success: boolean; data?: Attachment; error?: string }> {
    try {
      const base = `/api/v1/projects/${projectKey}/tasks/${taskNum}`;
      const url =
        entityType === "task"
          ? `${base}/attachments`
          : `${base}/comments/${commentId}/attachments`;

      const response = await fetch(url, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        body: JSON.stringify({ upload_id: uploadId }),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message };
      }

      const data: Attachment = await response.json();
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function deleteAttachment(
    projectKey: string,
    taskNum: number,
    entityType: "task" | "comment",
    attachmentId: string,
    commentId?: string
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const base = `/api/v1/projects/${projectKey}/tasks/${taskNum}`;
      const url =
        entityType === "task"
          ? `${base}/attachments/${attachmentId}`
          : `${base}/comments/${commentId}/attachments/${attachmentId}`;

      const response = await fetch(url, {
        method: "DELETE",
        headers: getAuthHeader(),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message };
      }

      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  return { listAttachments, attachFile, deleteAttachment };
}

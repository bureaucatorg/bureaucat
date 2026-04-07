interface UploadResponse {
  id: string;
  filename: string;
  mime_type: string;
  size_bytes: number;
  url: string;
}

export function useUploads() {
  const { getAuthHeader } = useAuth();

  async function uploadFile(
    file: File
  ): Promise<{ success: boolean; data?: UploadResponse; error?: string }> {
    try {
      const formData = new FormData();
      formData.append("file", file);

      const response = await fetch("/api/v1/uploads", {
        method: "POST",
        headers: getAuthHeader(),
        body: formData,
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to upload file" };
      }

      const data: UploadResponse = await response.json();
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  function getUploadUrl(uploadId: string): string {
    return `/api/v1/uploads/${uploadId}`;
  }

  return {
    uploadFile,
    getUploadUrl,
  };
}

interface UploadedFile {
  uploadId: string;
  filename: string;
  mimeType: string;
  url: string;
}

export function useFileAttach() {
  const { uploadFile } = useUploads();

  const uploading = ref(false);

  async function uploadFiles(files: File[]): Promise<UploadedFile[]> {
    uploading.value = true;
    const results: UploadedFile[] = [];

    try {
      for (const file of files) {
        const res = await uploadFile(file);
        if (res.success && res.data) {
          results.push({
            uploadId: res.data.id,
            filename: file.name,
            mimeType: file.type,
            url: res.data.url,
          });
        }
      }
    } finally {
      uploading.value = false;
    }

    return results;
  }

  return { uploadFiles, uploading };
}

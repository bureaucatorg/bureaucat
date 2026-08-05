export interface UploadedFile {
  uploadId: string;
  filename: string;
  mimeType: string;
  url: string;
}

import { toast } from "vue-sonner";

// Long filenames are single unbreakable tokens; shorten them so a toast or
// label keeps its shape.
function shortName(name: string, max = 36): string {
  if (name.length <= max) return name;
  return `${name.slice(0, max - 13)}…${name.slice(-12)}`;
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
        } else {
          // Rejections (oversized file, unsupported type, ...) are reported
          // here so every caller gets feedback instead of failing silently.
          toast.error(`${shortName(file.name)}: ${res.error || "upload failed"}`);
        }
      }
    } finally {
      uploading.value = false;
    }

    return results;
  }

  return { uploadFiles, uploading };
}

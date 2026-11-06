export type PreviewKind = "pdf" | "csv" | "markdown" | "text";

export function previewKind(file: { filename: string; mime_type: string }): PreviewKind | null {
  const ext = file.filename.split(".").pop()?.toLowerCase();
  const mime = file.mime_type.split(";")[0]?.trim().toLowerCase();
  if (ext === "pdf" || mime === "application/pdf") return "pdf";
  if (ext === "csv" || mime === "text/csv") return "csv";
  if (ext === "md" || ext === "markdown" || mime === "text/markdown") return "markdown";
  if (ext === "txt" || mime === "text/plain") return "text";
  return null;
}

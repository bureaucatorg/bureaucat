import { marked } from "marked";

const renderer = new marked.Renderer();
renderer.link = ({ href, title, text }) => {
  const titleAttr = title ? ` title="${title}"` : "";
  return `<a href="${href}"${titleAttr} target="_blank" rel="noopener noreferrer">${text}</a>`;
};
marked.setOptions({ breaks: true, gfm: true, renderer });

/**
 * Convert a description string to HTML for the Tiptap editor.
 *
 * Descriptions may be stored as HTML (produced by the Tiptap editor) or as
 * legacy markdown (older tasks/templates authored via plain textareas). HTML is
 * passed through untouched; anything else is parsed as markdown. The `<` guard
 * mirrors the check used in the task detail view.
 */
export function mdToHtml(value: string | null | undefined): string {
  const str = value ?? "";
  if (!str) return "";
  return str.startsWith("<") ? str : (marked(str) as string);
}

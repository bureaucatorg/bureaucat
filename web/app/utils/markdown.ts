import DOMPurify from "dompurify";
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

// Every user-authored HTML string must pass through this before reaching v-html.
export function sanitizeHtml(html: string): string {
  return DOMPurify.sanitize(html, { ADD_ATTR: ["target"] });
}

export function renderRichText(value: string | null | undefined): string {
  return sanitizeHtml(mdToHtml(value));
}

export function renderMarkdown(value: string | null | undefined): string {
  return sanitizeHtml(marked(value ?? "") as string);
}

const ASSISTANT_FORBID_TAGS = ["img", "picture", "source", "video", "audio", "track", "iframe", "object", "embed", "svg", "math", "style", "link", "meta", "form", "input", "button"];

// Assistant replies can echo text planted by other users, so nothing may load remote
// resources or link off-site: media is dropped and only same-app links stay clickable.
export function renderAssistantMarkdown(value: string | null | undefined): string {
  const html = DOMPurify.sanitize(marked(value ?? "") as string, {
    FORBID_TAGS: ASSISTANT_FORBID_TAGS,
    FORBID_ATTR: ["style", "srcset", "background", "poster", "action", "formaction"],
  });
  const template = document.createElement("template");
  template.innerHTML = html;
  for (const a of template.content.querySelectorAll("a")) {
    const href = a.getAttribute("href") ?? "";
    if (/^\/(?![/\\])/.test(href)) {
      a.removeAttribute("target");
      a.removeAttribute("rel");
    } else {
      a.replaceWith(document.createTextNode(a.textContent ?? ""));
    }
  }
  return template.innerHTML;
}

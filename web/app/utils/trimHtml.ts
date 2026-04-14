/**
 * Trim leading/trailing whitespace from Tiptap-style HTML content.
 * - Removes empty top-level paragraphs (`<p></p>`, `<p><br></p>`) at the start and end.
 * - Trims whitespace and <br> tags on the inside edge of the first/last paragraph.
 * - Returns "" when the content is visually empty (so callers can treat that as "no content").
 */
export function trimHtmlContent(html: string): string {
  if (!html) return "";
  let out = html;
  const emptyPara = /^\s*<p>(?:\s|&nbsp;|<br\s*\/?>)*<\/p>\s*/i;
  const trailingEmptyPara = /\s*<p>(?:\s|&nbsp;|<br\s*\/?>)*<\/p>\s*$/i;
  // Strip fully-empty paragraphs from both ends.
  while (emptyPara.test(out)) out = out.replace(emptyPara, "");
  while (trailingEmptyPara.test(out)) out = out.replace(trailingEmptyPara, "");
  // Trim whitespace / <br> inside the first and last paragraph.
  out = out.replace(
    /^(<p[^>]*>)((?:\s|&nbsp;|<br\s*\/?>)+)/i,
    "$1"
  );
  out = out.replace(
    /((?:\s|&nbsp;|<br\s*\/?>)+)(<\/p>)(?!.*<\/p>)/is,
    "$2"
  );
  out = out.trim();
  // Collapse to empty if nothing meaningful remains.
  const textOnly = out.replace(/<[^>]*>/g, "").replace(/&nbsp;/g, " ").trim();
  return textOnly.length === 0 ? "" : out;
}

// nginx rewrites the placeholder from APP_URL at serve time; dev falls back to VITE_APP_URL.
// The prerender keeps the placeholder in hrefs so nginx fills those in too.
const PLACEHOLDER = "__APP_URL__";

function resolve() {
  if (import.meta.env.SSR) return PLACEHOLDER;
  const fromMeta = document.querySelector<HTMLMetaElement>('meta[name="app-url"]')?.content ?? "";
  const raw = fromMeta.includes(PLACEHOLDER) ? (import.meta.env.VITE_APP_URL ?? "") : fromMeta;
  return /^https?:\/\//.test(raw) ? raw.replace(/\/+$/, "") : "";
}

export const APP_URL = resolve();

export const appLink = (path: string) => `${APP_URL}${path}`;

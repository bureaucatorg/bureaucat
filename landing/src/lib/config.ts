// nginx rewrites the meta tag's placeholder from APP_URL at serve time; dev falls back to VITE_APP_URL.
const fromMeta = document.querySelector<HTMLMetaElement>('meta[name="app-url"]')?.content ?? "";
const raw = fromMeta.includes("__APP_URL__") ? (import.meta.env.VITE_APP_URL ?? "") : fromMeta;

export const APP_URL = /^https?:\/\//.test(raw) ? raw.replace(/\/+$/, "") : "";

export const appLink = (path: string) => `${APP_URL}${path}`;

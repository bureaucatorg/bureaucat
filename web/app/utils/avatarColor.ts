// Palette of Tailwind bg+text color pairs used for deterministic avatar fallbacks.
// Ultra-light backgrounds with darker text for a soft, friendly look.
const AVATAR_COLORS = [
  "bg-red-100 text-red-800",
  "bg-orange-100 text-orange-800",
  "bg-amber-100 text-amber-800",
  "bg-yellow-100 text-yellow-800",
  "bg-lime-100 text-lime-800",
  "bg-green-100 text-green-800",
  "bg-emerald-100 text-emerald-800",
  "bg-teal-100 text-teal-800",
  "bg-cyan-100 text-cyan-800",
  "bg-sky-100 text-sky-800",
  "bg-blue-100 text-blue-800",
  "bg-indigo-100 text-indigo-800",
  "bg-violet-100 text-violet-800",
  "bg-purple-100 text-purple-800",
  "bg-fuchsia-100 text-fuchsia-800",
  "bg-pink-100 text-pink-800",
  "bg-rose-100 text-rose-800",
] as const;

/**
 * Returns a deterministic Tailwind bg+text class string for a given seed
 * (typically a user id). The same seed always yields the same color.
 */
export function getAvatarColor(seed: string | number | null | undefined): string {
  if (seed === null || seed === undefined || seed === "") {
    return "bg-muted";
  }
  const s = String(seed);
  // Simple 32-bit FNV-ish hash — good enough for palette selection.
  let hash = 0;
  for (let i = 0; i < s.length; i++) {
    hash = (hash * 31 + s.charCodeAt(i)) | 0;
  }
  const idx = Math.abs(hash) % AVATAR_COLORS.length;
  return AVATAR_COLORS[idx]!;
}

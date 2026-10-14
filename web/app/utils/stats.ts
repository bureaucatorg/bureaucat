import { type DateValue, today, getLocalTimeZone } from "@internationalized/date";

export const STATS_PRESETS = [
  { label: "Last 7 days", days: 7 },
  { label: "Last 14 days", days: 14 },
  { label: "Last 30 days", days: 30 },
  { label: "Last 90 days", days: 90 },
];

const MONTHS = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];

export interface TrendSeries {
  key: string;
  label: string;
  color: string;
}

export interface StatsBar {
  label: string;
  count: number;
  fill: string;
}

export function statsToday(): DateValue {
  return today(getLocalTimeZone());
}

export function toIsoDay(d: DateValue): string {
  const p = (n: number) => String(n).padStart(2, "0");
  return `${d.year}-${p(d.month)}-${p(d.day)}`;
}

function fmtDate(d: DateValue): string {
  return `${MONTHS[d.month - 1]} ${d.day}, ${d.year}`;
}

// Inclusive day span between two dates.
export function spanDays(from: DateValue, to: DateValue): number {
  const tz = getLocalTimeZone();
  return Math.round((to.toDate(tz).getTime() - from.toDate(tz).getTime()) / 86_400_000) + 1;
}

// The preset the range matches (ending today), if any.
export function activePresetDays(from: DateValue, to: DateValue): number | null {
  if (toIsoDay(to) !== toIsoDay(statsToday())) return null;
  return STATS_PRESETS.find((p) => p.days === spanDays(from, to))?.days ?? null;
}

export function statsRangeLabel(from: DateValue, to: DateValue): string {
  const preset = STATS_PRESETS.find((p) => p.days === activePresetDays(from, to));
  return preset ? preset.label : `${fmtDate(from)} – ${fmtDate(to)}`;
}

// Parse "YYYY-MM-DD" without timezone drift and render a short label.
export function shortDay(day: string | undefined): string {
  if (!day) return "";
  const [, m, d] = day.split("-").map(Number);
  return `${MONTHS[(m ?? 1) - 1]} ${d}`;
}

function compactNum(v: number): string {
  if (Math.abs(v) >= 1000) {
    const n = v / 1000;
    return `${Number.isInteger(n) ? n : n.toFixed(1)}k`;
  }
  return v.toString();
}

// Integer-only y ticks so small ranges don't render duplicate fractional labels.
export function yTick(v: number): string {
  return Number.isInteger(v) ? compactNum(v) : "";
}

// Chart tooltips receive the bar index; these map it back to a readable heading.
export function dayLabelFor(data: { day: string }[]) {
  return (i: number | Date) => shortDay(data[Math.round(Number(i))]?.day);
}

export function categoryLabelFor(bars: { label: string }[]) {
  return (i: number | Date) => bars[Math.round(Number(i))]?.label ?? "";
}

// Human-readable byte size (1536 -> "1.5 KB").
export function formatBytes(bytes: number): string {
  if (bytes <= 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  const value = bytes / 1024 ** i;
  return `${i === 0 ? value : value.toFixed(1)} ${units[i]}`;
}

export function titleCase(value: string): string {
  return value.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}
